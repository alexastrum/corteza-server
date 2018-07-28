package service

import (
	"context"
	"github.com/crusttech/crust/sam/repository"
	"github.com/crusttech/crust/sam/types"
)

type (
	team struct {
		repository teamRepository
	}

	teamRepository interface {
		FindByID(ctx context.Context, teamID uint64) (*types.Team, error)
		Find(ctx context.Context, filter *types.TeamFilter) ([]*types.Team, error)

		Create(ctx context.Context, team *types.Team) (*types.Team, error)
		Update(ctx context.Context, team *types.Team) (*types.Team, error)

		Merge(ctx context.Context, teamID, targetTeamID uint64) error
		Move(ctx context.Context, teamID, organisationID uint64) error

		deleter
		archiver
	}
)

func Team() *team {
	return &team{repository: repository.Team()}
}

func (svc team) FindByID(ctx context.Context, id uint64) (*types.Team, error) {
	// @todo: permission check if current user has access to this team
	return svc.repository.FindByID(ctx, id)
}

func (svc team) Find(ctx context.Context, filter *types.TeamFilter) ([]*types.Team, error) {
	// @todo: permission check to return only teams that current user has access to
	return svc.repository.Find(ctx, filter)
}

func (svc team) Create(ctx context.Context, mod *types.Team) (*types.Team, error) {
	// @todo: permission check if current user can add/edit team

	return svc.repository.Create(ctx, mod)
}

func (svc team) Update(ctx context.Context, mod *types.Team) (*types.Team, error) {
	// @todo: permission check if current user can add/edit team
	// @todo: make sure archived & deleted entries can not be edited

	return svc.repository.Update(ctx, mod)
}

func (svc team) Delete(ctx context.Context, id uint64) error {
	// @todo: make history unavailable
	// @todo: notify users that team has been removed (remove from web UI)
	// @todo: permissions check if current user can remove team
	return svc.repository.Delete(ctx, id)
}

func (svc team) Archive(ctx context.Context, id uint64) error {
	// @todo: make history unavailable
	// @todo: notify users that team has been removed (remove from web UI)
	// @todo: permissions check if current user can remove team
	return svc.repository.Archive(ctx, id)
}

func (svc team) Unarchive(ctx context.Context, id uint64) error {
	// @todo: permissions check if current user can unarchive team
	// @todo: make history accessible
	// @todo: notify users that team has been unarchived
	return svc.repository.Unarchive(ctx, id)
}

func (svc team) Merge(ctx context.Context, id, targetTeamID uint64) error {
	// @todo: permission check if current user can merge team
	return svc.repository.Merge(ctx, id, targetTeamID)
}

func (svc team) Move(ctx context.Context, id, targetOrganisationID uint64) error {
	// @todo: permission check if current user can move team to another organisation
	return svc.repository.Move(ctx, id, targetOrganisationID)
}