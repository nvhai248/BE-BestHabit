package gintask

import (
	"bestHabit/common"
	"bestHabit/component"
	"bestHabit/modules/task/taskbiz"
	"bestHabit/modules/task/taskmodel"
	"bestHabit/modules/task/taskstorage"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Get List User's Task
// @Description User Get List User's Task after successful authentication.
// @Tags Tasks
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Authorization"\
// @Param name path string true "Task's name"
// @Param page path number true "Page number"
// @Param limit path number true "Limit of tasks returned!"
// @Param cursor path string true "Task Id"
// @Param deadline path string true "Deadline"
// @Success 200 {object} []taskmodel.Task "Successfully!"
// @Router /api/tasks [get]
func ListTaskByConditions(appCtx component.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter taskmodel.TaskFilter

		// Bind query parameters to the filter struct
		if err := c.ShouldBind(&filter); err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		var paging common.Paging

		// Bind query parameters to the filter struct
		if err := c.ShouldBind(&paging); err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		paging.Fulfill()

		var conditions common.Conditions

		// Bind conditions from
		if err := c.ShouldBind(&conditions); err != nil {
			panic(common.ErrInvalidRequest(err))
		}

		mapConditions := common.ConvertToMap(conditions)

		mapConditions["user_id"] = c.MustGet(common.CurrentUser).(common.Requester).GetId()

		store := taskstorage.NewSQLStore(appCtx.GetMainDBConnection())
		biz := taskbiz.NewListTaskBiz(store)

		result, err := biz.ListTask(c.Request.Context(), &filter, &paging, mapConditions)

		if err != nil {
			panic(err)
		}

		for i := range result {
			result[i].Mask(false)

			if i == len(result)-1 {
				paging.NextCursor = result[i].FakeID.String()
			}
		}

		c.JSON(http.StatusOK, common.NewSuccessResponse(result, paging, filter, 200, "Successful!"))
	}
}
