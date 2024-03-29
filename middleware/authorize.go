package middleware

import (
	"bestHabit/common"
	"bestHabit/component"
	"bestHabit/component/tokenprovider/jwt"
	"bestHabit/modules/user/userstorage"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func ErrWrongAuthHeader(err error) *common.AppError {
	return common.NewCustomError(
		err,
		fmt.Sprintf("wrong auth header"),
		fmt.Sprintf("ErrWrongAuthHeader"),
	)
}

func extractTokenFromHeaderString(s string) (string, error) {
	parts := strings.Split(s, " ")
	// "Authorization" : "Bearer {token}"

	if parts[0] != "Bearer" || len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return "", ErrWrongAuthHeader(nil)
	}

	return parts[1], nil
}

// 1. Get token from header
// 2. Validate token and parse to payload
// 3. From the token payload, we use user_id to find from DB

func RequireAuth(appCtx component.AppContext) func(c *gin.Context) {

	tokenProvider := jwt.NewTokenJWTProvider(appCtx.SecretKey())

	return func(c *gin.Context) {
		token, err := extractTokenFromHeaderString(c.GetHeader("Authorization"))

		if err != nil {
			panic(err)
		}

		db := appCtx.GetMainDBConnection()
		store := userstorage.NewSQLStore(db)

		payload, err := tokenProvider.Validate(token)
		if err != nil {
			panic(err)
		}

		user, err := store.FindById(c.Request.Context(), payload.UserId)

		if err != nil {
			panic(err)
		}

		if user.Status == common.UserDeleted {
			panic(common.ErrNoPermission(errors.New("user has been deleted!")))
		}

		if user.Status == common.UserBanned {
			panic(common.ErrNoPermission(errors.New("user has been banned!")))
		}

		user.Mask(false)

		// save user in context
		c.Set(common.CurrentUser, user)
		c.Next()
	}
}

func CompareIdBeforeVerify(appCtx component.AppContext) func(c *gin.Context) {
	tokenProvider := jwt.NewTokenJWTProvider(appCtx.SecretKey())

	return func(c *gin.Context) {

		token := c.Param("token")

		payload, err := tokenProvider.Validate(token)
		if err != nil {
			panic(common.NewCustomError(nil, "Wrong token to compare!", "ErrWrongAuth"))
		}

		user := c.MustGet(common.CurrentUser).(common.Requester)

		if user.GetId() != payload.UserId {
			panic(common.NewCustomError(nil, "You are not have permission to verify!", "NoPermission"))
		}

		c.Next()
	}
}
