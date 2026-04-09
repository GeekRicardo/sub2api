package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) ListRequestDetails(ctx context.Context, filter *service.OpsRequestDetailFilter) ([]*service.OpsRequestDetail, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil ops repository")
	}

	page, pageSize, startTime, endTime := filter.Normalize()
	offset := (page - 1) * pageSize

	conditions := make([]string, 0, 16)
	args := make([]any, 0, 24)

	// Placeholders $1/$2 reserved for time window inside the CTE.
	args = append(args, startTime.UTC(), endTime.UTC())

	addCondition := func(condition string, values ...any) {
		conditions = append(conditions, condition)
		args = append(args, values...)
	}

	if filter != nil {
		if kind := strings.TrimSpace(strings.ToLower(filter.Kind)); kind != "" && kind != "all" {
			if kind != string(service.OpsRequestKindSuccess) && kind != string(service.OpsRequestKindError) {
				return nil, 0, fmt.Errorf("invalid kind")
			}
			addCondition(fmt.Sprintf("kind = $%d", len(args)+1), kind)
		}

		if platform := strings.TrimSpace(strings.ToLower(filter.Platform)); platform != "" {
			addCondition(fmt.Sprintf("platform = $%d", len(args)+1), platform)
		}
		if filter.GroupID != nil && *filter.GroupID > 0 {
			addCondition(fmt.Sprintf("group_id = $%d", len(args)+1), *filter.GroupID)
		}

		if filter.UserID != nil && *filter.UserID > 0 {
			addCondition(fmt.Sprintf("user_id = $%d", len(args)+1), *filter.UserID)
		}
		if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
			addCondition(fmt.Sprintf("api_key_id = $%d", len(args)+1), *filter.APIKeyID)
		}
		if filter.AccountID != nil && *filter.AccountID > 0 {
			addCondition(fmt.Sprintf("account_id = $%d", len(args)+1), *filter.AccountID)
		}

		if model := strings.TrimSpace(filter.Model); model != "" {
			addCondition(fmt.Sprintf("model = $%d", len(args)+1), model)
		}
		if requestID := strings.TrimSpace(filter.RequestID); requestID != "" {
			addCondition(fmt.Sprintf("request_id = $%d", len(args)+1), requestID)
		}
		if q := strings.TrimSpace(filter.Query); q != "" {
			like := "%" + strings.ToLower(q) + "%"
			startIdx := len(args) + 1
			addCondition(
				fmt.Sprintf("(LOWER(COALESCE(request_id,'')) LIKE $%d OR LOWER(COALESCE(model,'')) LIKE $%d OR LOWER(COALESCE(message,'')) LIKE $%d)",
					startIdx, startIdx+1, startIdx+2,
				),
				like, like, like,
			)
		}

		if filter.MinDurationMs != nil {
			addCondition(fmt.Sprintf("duration_ms >= $%d", len(args)+1), *filter.MinDurationMs)
		}
		if filter.MaxDurationMs != nil {
			addCondition(fmt.Sprintf("duration_ms <= $%d", len(args)+1), *filter.MaxDurationMs)
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	cte := `
WITH combined AS (
  SELECT
    ul.id AS detail_id,
    'success'::TEXT AS kind,
    ul.created_at AS created_at,
    ul.request_id AS request_id,
    COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    ul.model AS model,
    ul.duration_ms AS duration_ms,
    NULL::INT AS status_code,
    NULL::BIGINT AS error_id,
    NULL::TEXT AS phase,
    NULL::TEXT AS severity,
    NULL::TEXT AS message,
    ul.user_id AS user_id,
    ul.api_key_id AS api_key_id,
    ul.account_id AS account_id,
    ul.group_id AS group_id,
    ul.stream AS stream
  FROM usage_logs ul
  LEFT JOIN groups g ON g.id = ul.group_id
  LEFT JOIN accounts a ON a.id = ul.account_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2

  UNION ALL

  SELECT
    o.id AS detail_id,
    'error'::TEXT AS kind,
    o.created_at AS created_at,
    COALESCE(NULLIF(o.request_id,''), NULLIF(o.client_request_id,''), '') AS request_id,
    COALESCE(NULLIF(o.platform, ''), NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
    o.model AS model,
    o.duration_ms AS duration_ms,
    o.status_code AS status_code,
    o.id AS error_id,
    o.error_phase AS phase,
    o.severity AS severity,
    o.error_message AS message,
    o.user_id AS user_id,
    o.api_key_id AS api_key_id,
    o.account_id AS account_id,
    o.group_id AS group_id,
    o.stream AS stream
  FROM ops_error_logs o
  LEFT JOIN groups g ON g.id = o.group_id
  LEFT JOIN accounts a ON a.id = o.account_id
  WHERE o.created_at >= $1 AND o.created_at < $2
    AND COALESCE(o.status_code, 0) >= 400
)
`

	countQuery := fmt.Sprintf(`%s SELECT COUNT(1) FROM combined %s`, cte, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			total = 0
		} else {
			return nil, 0, err
		}
	}

	sort := "ORDER BY created_at DESC"
	if filter != nil {
		switch strings.TrimSpace(strings.ToLower(filter.Sort)) {
		case "", "created_at_desc":
			// default
		case "duration_desc":
			sort = "ORDER BY duration_ms DESC NULLS LAST, created_at DESC"
		default:
			return nil, 0, fmt.Errorf("invalid sort")
		}
	}

	listQuery := fmt.Sprintf(`
%s
SELECT
  detail_id,
  kind,
  created_at,
  request_id,
  platform,
  model,
  duration_ms,
  status_code,
  error_id,
  phase,
  severity,
  message,
  user_id,
  api_key_id,
  account_id,
  group_id,
  stream
FROM combined
%s
%s
LIMIT $%d OFFSET $%d
`, cte, where, sort, len(args)+1, len(args)+2)

	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	toIntPtr := func(v sql.NullInt64) *int {
		if !v.Valid {
			return nil
		}
		i := int(v.Int64)
		return &i
	}
	toInt64Ptr := func(v sql.NullInt64) *int64 {
		if !v.Valid {
			return nil
		}
		i := v.Int64
		return &i
	}

	out := make([]*service.OpsRequestDetail, 0, pageSize)
	for rows.Next() {
		var (
			detailID  int64
			kind      string
			createdAt time.Time
			requestID sql.NullString
			platform  sql.NullString
			model     sql.NullString

			durationMs sql.NullInt64
			statusCode sql.NullInt64
			errorID    sql.NullInt64

			phase    sql.NullString
			severity sql.NullString
			message  sql.NullString

			userID    sql.NullInt64
			apiKeyID  sql.NullInt64
			accountID sql.NullInt64
			groupID   sql.NullInt64

			stream bool
		)

		if err := rows.Scan(
			&detailID,
			&kind,
			&createdAt,
			&requestID,
			&platform,
			&model,
			&durationMs,
			&statusCode,
			&errorID,
			&phase,
			&severity,
			&message,
			&userID,
			&apiKeyID,
			&accountID,
			&groupID,
			&stream,
		); err != nil {
			return nil, 0, err
		}

		item := &service.OpsRequestDetail{
			DetailID:  detailID,
			Kind:      service.OpsRequestKind(kind),
			CreatedAt: createdAt,
			RequestID: strings.TrimSpace(requestID.String),
			Platform:  strings.TrimSpace(platform.String),
			Model:     strings.TrimSpace(model.String),

			DurationMs: toIntPtr(durationMs),
			StatusCode: toIntPtr(statusCode),
			ErrorID:    toInt64Ptr(errorID),
			Phase:      phase.String,
			Severity:   severity.String,
			Message:    message.String,

			UserID:    toInt64Ptr(userID),
			APIKeyID:  toInt64Ptr(apiKeyID),
			AccountID: toInt64Ptr(accountID),
			GroupID:   toInt64Ptr(groupID),

			Stream: stream,
		}

		if item.Platform == "" {
			item.Platform = "unknown"
		}

		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *opsRepository) GetSuccessRequestDetail(ctx context.Context, id int64) (*service.OpsSuccessRequestDetail, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}

	const q = `
SELECT
  ul.id,
  ul.created_at,
  COALESCE(ul.request_id, ''),
  COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), '') AS platform,
  COALESCE(ul.model, ''),
  COALESCE(ul.requested_model, ''),
  COALESCE(ul.upstream_model, ''),
  ul.duration_ms,
  ul.stream,
  ul.request_type,
  ul.user_id,
  COALESCE(u.email, ''),
  ul.api_key_id,
  ul.account_id,
  COALESCE(a.name, ''),
  ul.group_id,
  COALESCE(g.name, ''),
  COALESCE(ul.user_agent, ''),
  COALESCE(ul.inbound_endpoint, ''),
  COALESCE(ul.upstream_endpoint, ''),
  COALESCE(ul.request_body, ''),
  ul.request_body_truncated,
  ul.request_body_bytes,
  COALESCE(ul.response_body, ''),
  ul.response_body_truncated,
  ul.response_body_bytes,
  ul.input_tokens,
  ul.output_tokens,
  ul.cache_creation_tokens,
  ul.cache_read_tokens,
  ul.total_cost,
  ul.actual_cost
FROM usage_logs ul
LEFT JOIN users u ON u.id = ul.user_id
LEFT JOIN accounts a ON a.id = ul.account_id
LEFT JOIN groups g ON g.id = ul.group_id
WHERE ul.id = $1
LIMIT 1`

	var (
		out service.OpsSuccessRequestDetail

		durationMs        sql.NullInt64
		requestType       sql.NullInt64
		userID            sql.NullInt64
		apiKeyID          sql.NullInt64
		accountID         sql.NullInt64
		groupID           sql.NullInt64
		requestBodyBytes  sql.NullInt64
		responseBodyBytes sql.NullInt64
	)

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&out.ID,
		&out.CreatedAt,
		&out.RequestID,
		&out.Platform,
		&out.Model,
		&out.RequestedModel,
		&out.UpstreamModel,
		&durationMs,
		&out.Stream,
		&requestType,
		&userID,
		&out.UserEmail,
		&apiKeyID,
		&accountID,
		&out.AccountName,
		&groupID,
		&out.GroupName,
		&out.UserAgent,
		&out.InboundEndpoint,
		&out.UpstreamEndpoint,
		&out.RequestBody,
		&out.RequestBodyTruncated,
		&requestBodyBytes,
		&out.ResponseBody,
		&out.ResponseBodyTruncated,
		&responseBodyBytes,
		&out.InputTokens,
		&out.OutputTokens,
		&out.CacheCreationTokens,
		&out.CacheReadTokens,
		&out.TotalCost,
		&out.ActualCost,
	)
	if err != nil {
		return nil, err
	}

	if out.Platform == "" {
		out.Platform = "unknown"
	}
	if out.RequestedModel == "" {
		out.RequestedModel = out.Model
	}
	if durationMs.Valid {
		value := int(durationMs.Int64)
		out.DurationMs = &value
	}
	if requestType.Valid {
		value := int16(requestType.Int64)
		out.RequestType = &value
	}
	if userID.Valid {
		value := userID.Int64
		out.UserID = &value
	}
	if apiKeyID.Valid {
		value := apiKeyID.Int64
		out.APIKeyID = &value
	}
	if accountID.Valid {
		value := accountID.Int64
		out.AccountID = &value
	}
	if groupID.Valid {
		value := groupID.Int64
		out.GroupID = &value
	}
	if requestBodyBytes.Valid {
		value := int(requestBodyBytes.Int64)
		out.RequestBodyBytes = &value
	}
	if responseBodyBytes.Valid {
		value := int(responseBodyBytes.Int64)
		out.ResponseBodyBytes = &value
	}

	out.RequestBody = strings.TrimSpace(out.RequestBody)
	out.ResponseBody = strings.TrimSpace(out.ResponseBody)
	return &out, nil
}
