package bulk

// This file is auto-generated.
//
// Template: pkg/store_interfaces_joined.gen.go.tpl
// Definitions:
//  - store/actionlog.yaml
//  - store/applications.yaml
//  - store/compose_charts.yaml
//  - store/compose_module_fields.yaml
//  - store/compose_modules.yaml
//  - store/compose_namespaces.yaml
//  - store/compose_pages.yaml
//  - store/credentials.yaml
//  - store/rbac_rules.yaml
//  - store/reminders.yaml
//  - store/roles.yaml
//  - store/settings.yaml
//  - store/system_attachments.yaml
//  - store/users.yaml
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated.
//

type (
	// Interface combines interfaces of all supported store interfaces
	storeGeneratedInterfaces interface {
		actionlogsStore
		applicationsStore
		composeChartsStore
		composeModuleFieldsStore
		composeModulesStore
		composeNamespacesStore
		composePagesStore
		credentialsStore
		rbacRulesStore
		remindersStore
		rolesStore
		settingsStore
		attachmentsStore
		usersStore
	}
)