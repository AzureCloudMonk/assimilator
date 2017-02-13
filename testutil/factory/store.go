package factory

import (
	"github.com/diyan/assimilator/db/store"
	"github.com/diyan/assimilator/models"
)

func (tf TestFactory) SaveOrganization(org models.Organization) {
	orgStore := store.NewOrganizationStore(tf.ctx)
	tf.noError(orgStore.SaveOrganization(org))
}

func (tf TestFactory) SaveOrganizationMember(orgMember models.OrganizationMember) {
	orgStore := store.NewOrganizationStore(tf.ctx)
	tf.noError(orgStore.SaveOrganizationMember(orgMember))
}

func (tf TestFactory) SaveTeam(team models.Team) {
	teamStore := store.NewTeamStore(tf.ctx)
	tf.noError(teamStore.SaveTeam(team))
}

func (tf TestFactory) SaveTeamMember(teamMember models.OrganizationMemberTeam) {
	teamStore := store.NewTeamStore(tf.ctx)
	tf.noError(teamStore.SaveMember(teamMember))
}

func (tf TestFactory) SaveProject(project models.Project) {
	projectStore := store.NewProjectStore(tf.ctx)
	tf.noError(projectStore.SaveProject(project))
}

func (tf TestFactory) SaveTags(tags ...*models.TagKey) {
	projectStore := store.NewProjectStore(tf.ctx)
	tf.noError(projectStore.SaveTags(tags...))
}
