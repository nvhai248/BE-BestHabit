package subscriber

import (
	"bestHabit/common"
	"bestHabit/component"
	"bestHabit/component/asyncjob"
	"bestHabit/pubsub"
	"context"
	"log"
)

type consumerJob struct {
	Title string
	Hld   func(ctx context.Context, message *pubsub.Message) error
}

type consumerEngine struct {
	appCtx component.AppContext
}

func NewEngine(appContext component.AppContext) *consumerEngine {
	return &consumerEngine{appCtx: appContext}
}

func (engine *consumerEngine) Start() error {
	engine.startSubTopic(
		common.TopicUserAddNewDvToken,
		true,
		RunCreateNewCronJobAfterUserAddDeviceToken(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserCreateNewTask,
		true,
		RunIncreaseTaskCountAfterUserCreateNewTask(engine.appCtx),
		RunCreateNewCronJobTaskAfterUserAddNewTask(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserCreateNewHabit,
		true,
		RunIncreaseHabitCountAfterUserCreateNewHabit(engine.appCtx),
		RunCreateNewCronJobHabitAfterUserAddNewHabit(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserDeleteTask,
		true,
		RunDecreaseTaskCountAfterUserDeleteTask(engine.appCtx),
		RunDeleteCronJobTaskAfterUserDeleteTask(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserDeleteHabit,
		true,
		RunDecreaseHabitCountAfterUserDeleteHabit(engine.appCtx),
		RunDeleteCronJobHabitAfterUserDeleteHabit(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserJoinChallenge,
		true,
		RunIncreaseChallengeCountAfterUserJoinChallenge(engine.appCtx),
		RunIncreaseUserJoinedCountAfterUserJoinedChallenge(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserCancelChallenge,
		true,
		RunDecreaseChallengeCountAfterUserCancelChallenge(engine.appCtx),
		RunDecreaseUserJoinedCountAfterUserCancelChallenge(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserUpdateTask,
		true,
		RunUpdateCronJobTaskAfterUserUpdateTask(engine.appCtx),
	)

	engine.startSubTopic(
		common.TopicUserUpdateHabit,
		true,
		RunUpdateCronJobHabitAfterUserUpdateHabit(engine.appCtx),
	)

	return nil
}

type GroupJob interface {
	Run(ctx context.Context) error
}

func (engine *consumerEngine) startSubTopic(topic pubsub.Topic, isConcurrent bool, consumerJobs ...consumerJob) error {
	c, _ := engine.appCtx.GetPubSub().Subscribe(context.Background(), topic)

	for _, item := range consumerJobs {
		log.Println("Setup consumer for: ", item.Title)
	}

	getJobHandler := func(job *consumerJob, message *pubsub.Message) asyncjob.JobHandler {
		return func(ctx context.Context) error {
			log.Println("running job for ", job.Title, ". Value: ", message.Data())
			return job.Hld(ctx, message)
		}
	}

	go func() {
		for {
			mgs := <-c

			jobHdlArr := make([]asyncjob.Job, len(consumerJobs))

			for i := range consumerJobs {
				jobHdl := getJobHandler(&consumerJobs[i], mgs)
				jobHdlArr[i] = asyncjob.NewJob(jobHdl)
			}

			group := asyncjob.NewGroup(isConcurrent, jobHdlArr...)

			if err := group.Run(context.Background()); err != nil {
				log.Println(err)
			}
		}
	}()

	return nil
}
