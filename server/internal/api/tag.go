package api

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/imdotdev/im.dev/server/internal/interaction"
	"github.com/imdotdev/im.dev/server/internal/notification"
	"github.com/imdotdev/im.dev/server/internal/story"
	"github.com/imdotdev/im.dev/server/internal/tags"
	"github.com/imdotdev/im.dev/server/internal/user"
	"github.com/imdotdev/im.dev/server/pkg/common"
	"github.com/imdotdev/im.dev/server/pkg/e"
	"github.com/imdotdev/im.dev/server/pkg/models"
)

func GetTag(c *gin.Context) {
	name := c.Param("name")
	res, err := tags.GetTag("", name)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

func GetTags(c *gin.Context) {
	filter := c.Query("filter")
	res, err := models.GetTags()
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	if filter == models.FilterFavorites {
		for _, tag := range res {
			tag.Follows = interaction.GetFollows(tag.ID)
		}

		sort.Sort(models.FollowTags(res))
	} else {
		sort.Sort(res)
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

func GetTagsByIDs(c *gin.Context) {
	ids := make([]string, 0)
	err := c.Bind(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}

	ts, err0 := tags.GetTagsByIDs(ids)
	if err != nil {
		c.JSON(err0.Status, common.RespError(err0.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(ts))
}

func SubmitTag(c *gin.Context) {
	tag := &models.Tag{}
	err := c.Bind(&tag)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}

	user := user.CurrentUser(c)
	if !tags.IsModerator(tag.ID, user) {
		c.JSON(http.StatusForbidden, common.RespError(e.NoEditorPermission))
		return
	}

	tag.Creator = user.ID
	err1 := tags.SubmitTag(tag)
	if err1 != nil {
		c.JSON(err1.Status, common.RespError(err1.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(nil))
}

func DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}

	user := user.CurrentUser(c)
	if !user.Role.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, common.RespError(e.NoPermission))
	}

	err := tags.DeleteTag(id)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(nil))
}

func GetUserTags(c *gin.Context) {
	userID := c.Param("userID")
	res, err := tags.GetUserTags(userID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

func GetOrgTags(c *gin.Context) {
	userID := c.Param("userID")
	res, err := tags.GetUserTags(userID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

func GetTagModerators(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}

	res, err := tags.GetModerators(id)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

type AddModeratorReq struct {
	TagID    string `json:"tagID"`
	Username string `json:"username"`
}

func AddModerator(c *gin.Context) {
	req := &AddModeratorReq{}
	c.Bind(&req)

	user := user.CurrentUser(c)
	if !user.Role.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, common.RespError(e.NoPermission))
		return
	}

	err := tags.AddModerator(req.TagID, req.Username)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(nil))
}

func DeleteModerator(c *gin.Context) {
	tagID := c.Param("tagID")
	userID := c.Param("userID")
	if tagID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}
	user := user.CurrentUser(c)
	if !user.Role.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, common.RespError(e.NoPermission))
	}

	err := tags.DeleteModerator(tagID, userID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(nil))
}

func RemoveTagStory(c *gin.Context) {
	tagID := c.Param("tagID")
	storyID := c.Param("storyID")
	if tagID == "" || storyID == "" {
		c.JSON(http.StatusBadRequest, common.RespError(e.ParamInvalid))
		return
	}

	user := user.CurrentUser(c)
	if !tags.IsModerator(tagID, user) {
		c.JSON(http.StatusForbidden, common.RespError(e.NoEditorPermission))
		return
	}

	err := tags.RemoveTagStory(tagID, storyID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	err = tags.DisableTagStory(tagID, storyID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	s, err := story.GetStory(storyID, "")
	if err == nil {
		t, err := tags.GetTag(tagID, "")
		if err == nil {
			notification.Send(s.CreatorID, "", models.NotificationSystem, storyID, " delete your story from tag "+t.Name, user.ID)
		}
	}

	c.JSON(http.StatusOK, common.RespSuccess(nil))

}

func GetTagListByUserModeratorRole(c *gin.Context) {
	user := user.CurrentUser(c)

	res, err := tags.GetTagListByUserModeratorRole(user.ID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}

func GetTagDisabledStroies(c *gin.Context) {
	tagID := c.Param("tagID")

	res, err := tags.GetDisabledStroies(tagID)
	if err != nil {
		c.JSON(err.Status, common.RespError(err.Message))
		return
	}

	c.JSON(http.StatusOK, common.RespSuccess(res))
}
