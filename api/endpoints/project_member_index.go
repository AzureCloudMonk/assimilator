package api

import (
	"net/http"

	"github.com/diyan/assimilator/context"
	"github.com/diyan/assimilator/models"
	"github.com/pkg/errors"
)

// User ..
type User struct {
	models.User
	AvatarURL string      `json:"avatarUrl"`
	Options   UserOptions `json:"options"`
}

// UserOptions ..
type UserOptions struct {
	Timezone        string `json:"timezone"`        // TODO double check this
	StacktraceOrder string `json:"stacktraceOrder"` // default
	Language        string `json:"language"`
	Clock24Hours    bool   `json:"clock24Hours"`
}

func ProjectMemberIndexGetEndpoint(c context.Project) error {
	// TODO not clear what this expr means -> Q(user__is_active=True) | Q(user__isnull=True)
	users := []*User{}
	_, err := c.Tx.SelectBySql(`
		select u.*
			from auth_user u
				join sentry_organizationmember om on u.id = om.user_id
				join sentry_organization o on om.organization_id = o.id
		where o.id = ? and u.is_active = true`,
		c.Project.OrganizationID).
		LoadStructs(&users)
	if err != nil {
		return errors.Wrap(err, "can not read project members")
	}
	for _, user := range users {
		user.PostGet()
		// TODO add real implementation
		user.Options.Language = "en"
		user.Options.Timezone = "UTC"
		user.Options.StacktraceOrder = "default"
	}
	// TODO fill user.AvatarURL, user.Options. Check UserSerializer(Serializer) impl
	return c.JSON(http.StatusOK, users)
}
