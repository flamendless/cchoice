-- name: CreateStaffLeaveCreditGrant :one
INSERT INTO tbl_staff_leave_credit_grants (
    staff_id,
    hours,
    note,
    granted_by,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: SumStaffLeaveCreditHours :one
SELECT COALESCE(SUM(hours), 0)
FROM tbl_staff_leave_credit_grants
WHERE staff_id = ?;
