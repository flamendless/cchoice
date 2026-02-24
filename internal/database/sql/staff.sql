-- name: GetStaffByEmail :one
SELECT
    id,
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    date_hired,
	time_in_schedule,
	time_out_schedule,
    position,
    user_type,
    email,
    mobile_no,
    password,
    created_at,
    updated_at
FROM tbl_staffs
WHERE
    email = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetStaffByID :one
SELECT
    id,
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    date_hired,
	time_in_schedule,
	time_out_schedule,
    position,
    user_type,
    email,
    mobile_no,
    password,
    created_at,
    updated_at
FROM tbl_staffs
WHERE
    id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetAllStaffs :many
SELECT
    id,
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    date_hired,
	time_in_schedule,
	time_out_schedule,
    position,
    user_type,
    email,
    mobile_no,
    created_at,
    updated_at
FROM tbl_staffs
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY last_name ASC, first_name ASC
LIMIT ?;

-- name: GetStaffAttendanceByDate :one
SELECT
    id,
    staff_id,
    for_date,
    time_in,
    time_out,
    location,
    created_at,
    updated_at
FROM tbl_staff_attendances
WHERE
    staff_id = ?
    AND for_date = ?
LIMIT 1;

-- name: GetStaffAttendanceByStaffIDAndDateRange :many
SELECT
    sa.id,
    sa.staff_id,
    sa.for_date,
    sa.time_in,
    sa.time_out,
    sa.location,
    sa.created_at,
    sa.updated_at,
    s.first_name,
    s.middle_name,
    s.last_name
FROM tbl_staff_attendances sa
INNER JOIN tbl_staffs s ON s.id = sa.staff_id
WHERE
    sa.for_date = ?
    AND s.deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY s.last_name ASC, s.first_name ASC;

-- name: CreateStaff :one
INSERT INTO tbl_staffs (
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    date_hired,
	time_in_schedule,
	time_out_schedule,
    position,
    user_type,
    email,
    mobile_no,
    password,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: CreateStaffAttendance :one
INSERT INTO tbl_staff_attendances (
    staff_id,
    for_date,
    time_in,
    time_out,
    location,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: UpdateStaffAttendanceTimeIn :one
UPDATE tbl_staff_attendances
SET
    time_in = ?,
    location = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffAttendanceTimeOut :one
UPDATE tbl_staff_attendances
SET
    time_out = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffAttendanceLocation :one
UPDATE tbl_staff_attendances
SET
    location = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;
