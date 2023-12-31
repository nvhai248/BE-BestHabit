package participantstore

import (
	"bestHabit/common"
	"bestHabit/modules/participant/participantmodel"
	"context"
	"fmt"
	"strings"
)

func replacePlaceholders(query string, args []interface{}) string {
	for _, arg := range args {
		strVal := fmt.Sprintf("'%v'", arg)
		query = strings.Replace(query, "?", strVal, 1)
	}

	return query
}

func (s *sqlStore) ListParticipantByConditions(ctx context.Context,
	paging *common.Paging,
) ([]participantmodel.Participant, error) {
	db := s.db

	args := []interface{}{}
	query := "SELECT * FROM participants WHERE status <> 'cancel'"
	countQuery := "SELECT COUNT(*) FROM participants WHERE status <> 'cancel'"

	var conditionsAndMore string

	var participants []participantmodel.Participant
	limit := paging.Limit

	// count paging
	var total int64
	countQuery = db.Rebind(countQuery + conditionsAndMore)
	countQuery = replacePlaceholders(countQuery, args)

	if err := db.Get(&total, countQuery); err != nil {
		return nil, common.ErrDB(err)
	}

	paging.Total = total

	// update paging
	if v := paging.FakeCursor; v != "" {
		if uid, err := common.FromBase58(v); err == nil {
			conditionsAndMore = conditionsAndMore + fmt.Sprintf(" AND id < %d ", int(uid.GetLocalID())) + "ORDER BY id DESC LIMIT ?"
			args = append(args, limit)
		}
	} else {
		offset := (paging.Page - 1) * paging.Limit

		conditionsAndMore = conditionsAndMore + " ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	query = db.Rebind(query + conditionsAndMore)
	if err := db.Select(&participants, query, args...); err != nil {
		return nil, common.ErrDB(err)
	}

	return participants, nil
}
