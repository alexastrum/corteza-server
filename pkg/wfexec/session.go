package wfexec

import (
	"context"
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/auth"
	"github.com/cortezaproject/corteza-server/pkg/errors"
	"github.com/cortezaproject/corteza-server/pkg/expr"
	"github.com/cortezaproject/corteza-server/pkg/id"
	"github.com/cortezaproject/corteza-server/pkg/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

type (
	Session struct {
		// Session identifier
		id uint64

		// steps graph
		g *Graph

		started time.Time

		// state channel (ie work queue)
		qState chan *State

		// error channel
		qErr chan error

		// locks concurrent executions
		execLock chan struct{}

		// delayed states (waiting for the right time)
		delayed map[uint64]*delayed

		// prompted
		prompted map[uint64]*prompted

		// how often we check for delayed states and how often idle stat is checked in Wait()
		workerInterval time.Duration

		// only one worker routine per session
		workerLock chan struct{}

		workerTicker *time.Ticker

		statusChange chan int

		// holds final result
		result *expr.Vars
		err    error

		mux *sync.RWMutex

		// debug logger
		log *zap.Logger

		eventHandler StateChangeHandler
	}

	StateChangeHandler func(SessionStatus, *State, *Session)

	sessionOpt func(*Session)

	Frame struct {
		CreatedAt time.Time     `json:"createdAt"`
		SessionID uint64        `json:"sessionID"`
		StateID   uint64        `json:"stateID"`
		Input     *expr.Vars    `json:"input"`
		Scope     *expr.Vars    `json:"scope"`
		ParentID  uint64        `json:"parentID"`
		StepID    uint64        `json:"stepID"`
		LeadTime  time.Duration `json:"leadTime"`
	}

	// ExecRequest is passed to Exec() functions and contains all information
	// for ste pexecution
	ExecRequest struct {
		SessionID uint64
		StateID   uint64

		// Current input received on session resume
		Input *expr.Vars

		// Current scope
		Scope *expr.Vars

		// Helps with gateway join/merge steps
		// that needs info about the step it's currently merging
		Parent Step
	}

	SessionStatus int
)

const (
	sessionStateChanBuf   = 512
	sessionConcurrentExec = 32
)

const (
	SessionActive SessionStatus = iota
	SessionPrompted
	SessionDelayed
	SessionFailed
	SessionCompleted
)

var (
	// wrapper around nextID that will aid service testing
	nextID = func() uint64 {
		return id.Next()
	}

	// wrapper around time.Now() that will aid service testing
	now = func() *time.Time {
		c := time.Now()
		return &c
	}
)

func (s SessionStatus) String() string {
	switch s {
	case SessionActive:
		return "active"
	case SessionPrompted:
		return "prompted"
	case SessionDelayed:
		return "delayed"
	case SessionFailed:
		return "failed"
	case SessionCompleted:
		return "completed"
	}

	return "UNKNOWN-SESSION-STATUS"
}

func NewSession(ctx context.Context, g *Graph, oo ...sessionOpt) *Session {
	s := &Session{
		g:        g,
		id:       nextID(),
		started:  *now(),
		qState:   make(chan *State, sessionStateChanBuf),
		qErr:     make(chan error, 1),
		execLock: make(chan struct{}, sessionConcurrentExec),
		delayed:  make(map[uint64]*delayed),
		prompted: make(map[uint64]*prompted),

		//workerInterval: time.Millisecond,
		workerInterval: time.Millisecond * 250, // debug mode rate
		workerLock:     make(chan struct{}, 1),

		mux: &sync.RWMutex{},

		log: zap.NewNop(),
		eventHandler: func(SessionStatus, *State, *Session) {
			// noop
		},
	}

	for _, o := range oo {
		o(s)
	}

	s.log = s.log.
		WithOptions(zap.AddStacktrace(zap.ErrorLevel)).
		With(zap.Uint64("sessionID", s.id))

	go s.worker(ctx)

	return s
}

func (s Session) Status() SessionStatus {
	defer s.mux.Unlock()
	s.mux.Lock()

	switch {
	case s.err != nil:
		return SessionFailed

	case len(s.prompted) > 0:
		return SessionPrompted

	case len(s.delayed) > 0:
		return SessionDelayed

	case s.result == nil:
		return SessionActive

	default:
		return SessionCompleted
	}

}

func (s Session) ID() uint64 { return s.id }

func (s Session) Idle() bool {
	return s.Status() != SessionActive
}

func (s *Session) Error() error {
	defer s.mux.RUnlock()
	s.mux.RLock()

	return s.err
}

func (s *Session) Result() *expr.Vars {
	defer s.mux.RUnlock()
	s.mux.RLock()

	return s.result
}

func (s *Session) Exec(ctx context.Context, step Step, scope *expr.Vars) error {
	if s.g.Len() == 0 {
		return fmt.Errorf("refusing to execute without steps")
	}

	if len(s.g.Parents(step)) > 0 {
		return fmt.Errorf("can not execute step with parents")
	}

	if scope == nil {

		scope, _ = expr.NewVars(nil)
	}

	return s.enqueue(ctx, NewState(s, auth.GetIdentityFromContext(ctx), nil, step, scope))
}

// Prompts fn returns all owner's pending prompts on this session
func (s *Session) PendingPrompts(ownerId uint64) (out []*PendingPrompt) {
	if ownerId == 0 {
		return
	}

	defer s.mux.RUnlock()
	s.mux.RLock()

	out = make([]*PendingPrompt, 0, len(s.prompted))

	for _, p := range s.prompted {
		if p.ownerId != ownerId {
			continue
		}

		pending := p.toPending()
		pending.SessionID = s.id
		out = append(out, pending)
	}

	return
}

func (s *Session) Resume(ctx context.Context, stateId uint64, input *expr.Vars) error {
	defer s.mux.Unlock()
	s.mux.Lock()

	var (
		i      = auth.GetIdentityFromContext(ctx)
		p, has = s.prompted[stateId]
	)
	if !has {
		return fmt.Errorf("unexisting state")
	}

	if i == nil || p.ownerId != i.Identity() {
		return fmt.Errorf("state access denied")
	}

	delete(s.prompted, stateId)

	// setting received input to state
	p.state.input = input

	return s.enqueue(ctx, p.state)
}

func (s *Session) enqueue(ctx context.Context, st *State) error {
	if st == nil {
		return fmt.Errorf("state is nil")
	}

	if st.step == nil {
		return fmt.Errorf("state step is nil")
	}

	if st.stateId == 0 {
		st.stateId = nextID()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()

	case s.qState <- st:
		s.log.Debug("add step to queue")
		return nil
	}
}

// does not wait for the whole wf to be complete but until:
//  - context timeout
//  - idle state
//  - error in error queue
func (s *Session) Wait(ctx context.Context) error {
	return s.WaitUntil(ctx, SessionFailed, SessionDelayed, SessionCompleted)
}

// WaitUntil blocks until workflow session gets into expected status
//
func (s *Session) WaitUntil(ctx context.Context, expected ...SessionStatus) error {
	indexed := make(map[SessionStatus]bool)
	for _, status := range expected {
		indexed[status] = true
	}

	if indexed[s.Status()] {
		return s.err
	}

	s.log.Debug(
		"waiting for status change",
		zap.Any("expecting", expected),
		zap.Duration("interval", s.workerInterval),
	)

	waitCheck := time.NewTicker(s.workerInterval)
	defer waitCheck.Stop()

	for {
		select {
		case <-waitCheck.C:
			if indexed[s.Status()] {
				s.log.Debug("waiting complete", zap.Stringer("status", s.Status()))
				// nothing in the pipeline
				return s.err
			}

		case <-ctx.Done():
			s.log.Debug("wait context done", zap.Error(ctx.Err()))
			return s.err
		}
	}
}

func (s *Session) worker(ctx context.Context) {
	defer s.Stop()

	// making sure
	defer close(s.workerLock)
	s.workerLock <- struct{}{}

	s.workerTicker = time.NewTicker(s.workerInterval)
	defer s.workerTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Debug("worker context done", zap.Error(ctx.Err()))
			return

		case <-s.workerTicker.C:
			s.log.Debug("checking for delayed states")
			s.queueScheduledSuspended()

		case st := <-s.qState:
			if st == nil {
				// stop worker
				s.log.Debug("worker done")
				return
			}

			s.log.Debug("pulled state from queue", zap.Uint64("stateID", st.stateId))
			if st.step == nil {
				s.log.Debug("done, stopping and setting results")
				defer s.mux.Unlock()
				s.mux.Lock()

				// making sure result != nil
				s.result = (&expr.Vars{}).Merge(st.scope)
				return
			}

			// add empty struct to chan to lock and to have control over numver of concurrent go processes
			// this will block if number of items in execLock chan reached value of sessionConcurrentExec
			s.execLock <- struct{}{}

			go func() {
				s.exec(ctx, st)
				st.completed = now()

				// remove single
				<-s.execLock

				status := s.Status()
				s.log.Debug("executed", zap.Uint64("stateID", st.stateId), zap.Stringer("status", status))
				// after exec lock is released call event handler with (new) session status
				s.eventHandler(status, st, s)
			}()

		case err := <-s.qErr:
			if err == nil {
				// stop worker
				return
			}

			s.log.Warn("worker completed with error", zap.Error(err))
			s.mux.Lock()
			s.err = err
			s.mux.Unlock()
			return
		}
	}
}

func (s *Session) Stop() {
	s.log.Debug("stopping session")
	defer s.workerTicker.Stop()
}

func (s Session) Suspended() bool {
	defer s.mux.RUnlock()
	s.mux.RLock()
	return len(s.delayed) > 0
}

func (s *Session) queueScheduledSuspended() {
	defer s.mux.Unlock()
	s.mux.Lock()

	for id, sus := range s.delayed {
		if !sus.resumeAt.IsZero() && sus.resumeAt.After(*now()) {
			continue
		}

		delete(s.delayed, id)

		// Set state input when step is resumed
		sus.state.input = expr.RVars{
			"resumed":  expr.Must(expr.NewBoolean(true)),
			"resumeAt": expr.Must(expr.NewDateTime(sus.resumeAt)),
		}.Vars()
		s.qState <- sus.state
	}
}

// executes single step, resolves response and schedule following steps for execution
func (s *Session) exec(ctx context.Context, st *State) {
	var (
		result ExecResponse
		scope  = (&expr.Vars{}).Merge(st.scope)
		next   Steps

		currLoop = st.loopCurr()

		log = s.log.With(zap.Uint64("stateID", st.stateId))
	)

	if st.step != nil {
		log = log.With(zap.Uint64("stepID", st.step.ID()))
	}

	// @todo enable this when not in debug mode
	//       - OR -
	//       find a way to stick stacktrace from panic to log
	//defer func() {
	//	reason := recover()
	//	if reason == nil {
	//		return
	//	}
	//
	//	switch reason := reason.(type) {
	//	case error:
	//		log.Error("workflow session crashed", zap.Error(reason))
	//		s.qErr <- fmt.Errorf("session %d step %d crashed: %w", s.id, st.step.ID(), reason)
	//	default:
	//		s.qErr <- fmt.Errorf("session %d step %d crashed: %v", s.id, st.step.ID(), reason)
	//	}
	//}()

	s.eventHandler(SessionActive, st, s)

	{
		if currLoop != nil && currLoop.Is(st.step) {
			result = currLoop
		} else {
			// push logger to context but raise the stacktrace level to panic
			// to prevent overly verbose traces
			ctx = logger.ContextWithValue(ctx, log.WithOptions(zap.AddStacktrace(zap.PanicLevel)))

			// Context received in exec() wil not have the identity we're expecting
			// so we need to pull it from state owner and add it to new context
			// that is set to step exec function
			ctxWithIdentity := auth.SetIdentityToContext(ctx, st.owner)

			result, st.err = st.step.Exec(ctxWithIdentity, st.MakeRequest())

			if iterator, isIterator := result.(Iterator); isIterator && st.err == nil {
				// Exec fn returned an iterator, adding loop to stack
				st.newLoop(iterator)
				if err := iterator.Start(ctx, scope); err != nil {
					s.qErr <- err
				}
			}
		}

		if st.err != nil {
			if st.errHandler != nil {
				// handling error with error handling
				// step set in one of the previous steps
				log.Warn("step execution error handled",
					zap.Uint64("errorHandlerStepId", st.errHandler.ID()),
					zap.Error(st.err),
				)

				_ = expr.Assign(scope, "error", expr.Must(expr.NewString(st.err.Error())))

				// copy error handler & disable it on state to prevent inf. loop
				// in case of another error in the error-handling branch
				eh := st.errHandler
				st.errHandler = nil
				if err := s.enqueue(ctx, st.Next(eh, scope)); err != nil {
					log.Warn("unable to queue", zap.Error(err))
				}

				return
			} else {
				if errors.IsAutomation(st.err) {
					s.qErr <- st.err
				} else {
					log.Error("step execution failed", zap.Error(st.err))
					s.qErr <- fmt.Errorf("session %d step %d execution failed: %w", s.id, st.step.ID(), st.err)
				}

				return
			}
		}

		switch l := result.(type) {
		case Iterator:
			// add looper to state
			var (
				err error
				n   Step
			)
			n, result, err = l.Next(ctx, scope)
			if err != nil {
				s.qErr <- err
				return
			}

			if n == nil {
				next = st.loopEnd()
			} else {
				next = Steps{n}
			}
		}

		log.Debug("step executed", zap.String("resultType", fmt.Sprintf("%T", result)))
		switch result := result.(type) {
		case *expr.Vars:
			// most common (successful) result
			// session will continue with configured child steps
			scope = scope.Merge(result)

			log.Debug("result variables", zap.Any("scope", result))
			result.Each(func(k string, v expr.TypedValue) error {
				log.Debug("result variables", zap.String("name", k), zap.Any("value", v))
				return nil
			})

		case *errHandler:
			// this step sets error handling step on current state
			// and continues on the current path
			st.errHandler = result.handler

			// find step that's not error handler and
			// use it for the next step
			for _, c := range s.g.Children(st.step) {
				if c != st.errHandler {
					next = Steps{c}
					break
				}
			}

		case *loopBreak:
			if currLoop == nil {
				s.qErr <- fmt.Errorf("session %d step %d break step not inside a loop", s.id, st.step.ID())
				return
			}

			// jump out of the loop
			next = st.loopEnd()
			log.Debug("breaking from iterator")

		case *loopContinue:
			if currLoop == nil {
				s.qErr <- fmt.Errorf("session %d step %d continue step not inside a loop", s.id, st.step.ID())
				return
			}

			// jump back to iterator
			next = Steps{currLoop.Iterator()}
			log.Debug("continuing with next iteration")

		case *partial:
			// *partial is returned when step needs to be executed again
			// it's used mainly for join gateway step that should be called multiple times (one for each parent path)
			return

		case *termination:
			// terminate all activities, all delayed tasks and exit right away
			log.Debug("termination", zap.Int("delayed", len(s.delayed)))
			s.delayed = nil
			s.qState <- FinalState(s, scope)
			return

		case *delayed:
			log.Debug("session delayed", zap.Time("at", result.resumeAt))

			result.state = st
			s.mux.Lock()
			s.delayed[st.stateId] = result
			s.mux.Unlock()
			return

		case *resumed:
			log.Debug("session resumed")

		case *prompted:
			if result.ownerId == 0 {
				s.qErr <- fmt.Errorf("session %d step %d without an owner", s.id, st.step.ID())
				return
			}

			result.state = st
			s.mux.Lock()
			s.prompted[st.stateId] = result
			s.mux.Unlock()
			return

		case Steps:
			// session continues with set of specified steps
			// steps MUST be configured in a graph as step's children
			next = result

		case Step:
			// session continues with a specified step
			// step MUST be configured in a graph as step's child
			next = Steps{result}

		default:
			s.qErr <- fmt.Errorf("session %d step %d unknown exec response type %T", s.id, st.step.ID(), result)
			return
		}
	}

	if len(next) == 0 {
		// step's exec did not return next steps (only gateway steps, iterators and loops controls usually do that)
		//
		// rely on graph and get next (children) steps from there
		next = s.g.Children(st.step)
	} else {
		// children returned from step's exec
		// do a quick sanity check
		cc := s.g.Children(st.step)
		if len(cc) > 0 && !cc.Contains(next...) {
			s.qErr <- fmt.Errorf("inconsistent relationship")
			return
		}
	}

	if currLoop != nil && len(next) == 0 {
		// gracefully handling last step of iteration branch
		// that does not point back to the iterator step
		next = Steps{currLoop.Iterator()}
		log.Debug("last step in iteration branch, going back", zap.Uint64("backStepId", next[0].ID()))
	}

	if len(next) == 0 {
		log.Debug("zero paths, finalizing")
		// using state to transport results and complete the worker loop
		s.qState <- FinalState(s, scope)
		return
	}

	for _, step := range next {
		log.Debug("next step queued", zap.Uint64("nextStepId", step.ID()))
		if err := s.enqueue(ctx, st.Next(step, scope)); err != nil {
			log.Error("unable to queue", zap.Error(err))
		}
	}

}

func SetWorkerInterval(i time.Duration) sessionOpt {
	return func(s *Session) {
		s.workerInterval = i
	}
}

func SetHandler(fn StateChangeHandler) sessionOpt {
	return func(s *Session) {
		s.eventHandler = fn
	}
}

func SetLogger(log *zap.Logger) sessionOpt {
	return func(s *Session) {
		s.log = log
	}
}

func (ss Steps) hash() map[Step]bool {
	out := make(map[Step]bool)
	for _, s := range ss {
		out[s] = true
	}

	return out
}

func (ss Steps) Contains(steps ...Step) bool {
	hash := ss.hash()
	for _, s1 := range steps {
		if !hash[s1] {
			return false
		}
	}

	return true
}
