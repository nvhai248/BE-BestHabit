package taskmodel

type TaskUpdate struct {
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Deadline    *string `json:"deadline" db:"deadline"`
	Reminder    *string `json:"reminder" db:"reminder"`
	Status      *string `json:"status" db:"status"`
	UserId      *int
	Id          *int
}

func (TaskUpdate) TableName() string {
	return Task{}.TableName()
}
