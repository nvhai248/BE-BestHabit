package subscriber

import (
	"bestHabit/common"
	"bestHabit/component"
	"bestHabit/modules/cronnoticehabit/cronnoticehabitmodel"
	"bestHabit/modules/cronnoticehabit/cronnoticehabitstorage"
	"bestHabit/modules/user/userstorage"
	"bestHabit/pubsub"
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
)

func RunUpdateCronJobHabitAfterUserUpdateHabit(appCtx component.AppContext) consumerJob {
	store := cronnoticehabitstorage.NewSQLStore(appCtx.GetMainDBConnection())
	userStore := userstorage.NewSQLStore(appCtx.GetMainDBConnection())

	return consumerJob{
		Title: "Update cron job after user update Habit!",
		Hld: func(ctx context.Context, message *pubsub.Message) error {
			// get data from message
			habitData := message.Data().(HasHabitInfoUpdate)

			// remove cron job and data in db
			cronJobs, err := store.GetListCronNoticeHabit(ctx, habitData.GetUserId(), habitData.GetHabitId())

			if err != nil {
				return nil
			}

			for _, cronJob := range cronJobs {
				appCtx.GetCronJob().RemoveJob(cronJob.EntryId)
			}

			store.DeleteCronNoticeHabits(ctx, habitData.GetUserId(), habitData.GetHabitId())

			userData, err := userStore.FindById(ctx, habitData.GetUserId())
			if err != nil {
				return nil
			}

			sendNoticeFunc := func(deviceToken string, title, body string) error {
				return appCtx.GetSendNotification().SendNotification(deviceToken, title, body)
			}

			for _, token := range *userData.DeviceTokens {
				// create new cron job
				entryIds, err := appCtx.GetCronJob().CreateNewJobs(*common.NewNotificationBasedHabit(habitData.GetUserId(),
					habitData.GetDescription(),
					habitData.GetName(),
					habitData.GetStartDate(),
					habitData.GetEndDate(),
					habitData.GetReminderTime(),
					*habitData.GetDays()), sendNoticeFunc,
					token.DeviceToken,
					fmt.Sprintf("Time to do %s", habitData.GetName()),
					habitData.GetDescription())

				if err != nil {
					fmt.Println(err)
					return err
				}

				// save to db
				for _, entryId := range entryIds {
					store.CreateNewCronNoticeHabit(ctx, &cronnoticehabitmodel.CronNoticeHabit{
						UserId:  habitData.GetUserId(),
						HabitId: habitData.GetHabitId(),
						EntryId: cron.EntryID(entryId),
					})
				}
			}

			return nil
		},
	}
}
