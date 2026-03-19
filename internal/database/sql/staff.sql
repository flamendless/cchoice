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
    require_in_shop,
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
    require_in_shop,
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
    require_in_shop,
    created_at,
    updated_at
FROM tbl_staffs
WHERE deleted_at = '1970-01-01 00:00:00+00:00' AND user_type = 'STAFF'
ORDER BY last_name ASC, first_name ASC
LIMIT ?;

-- name: GetAllStaffsForAdmin :many
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
    require_in_shop,
    created_at,
    updated_at
FROM tbl_staffs
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
AND (
    @search IS NULL
    OR @search = ''
    OR LOWER(first_name || ' ' || COALESCE(middle_name, '') || ' ' || last_name) LIKE '%' || LOWER(@search) || '%'
    OR LOWER(last_name || ', ' || first_name || ' ' || COALESCE(middle_name, '')) LIKE '%' || LOWER(@search) || '%'
)
ORDER BY last_name ASC, first_name ASC;

-- name: GetStaffAttendanceByDate :one
SELECT
    id,
    staff_id,
    for_date,
    time_in,
    time_out,
    in_location,
    out_location,
    in_useragent_id,
    out_useragent_id,
    lunch_break_in,
    lunch_break_out,
    lunch_break_in_location,
    lunch_break_out_location,
    lunch_break_in_useragent_id,
    lunch_break_out_useragent_id,
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
    sa.in_location,
    sa.out_location,
    sa.in_useragent_id,
    sa.out_useragent_id,
    sa.lunch_break_in,
    sa.lunch_break_out,
    sa.lunch_break_in_location,
    sa.lunch_break_out_location,
    sa.lunch_break_in_useragent_id,
    sa.lunch_break_out_useragent_id,
    sa.created_at,
    sa.updated_at,
    s.first_name,
    s.middle_name,
    s.last_name,
    in_ua.browser as in_browser,
    in_ua.browser_version as in_browser_version,
    in_ua.os as in_os,
    in_ua.device as in_device,
    out_ua.browser as out_browser,
    out_ua.browser_version as out_browser_version,
    out_ua.os as out_os,
    out_ua.device as out_device,
    lunch_break_in_ua.browser as lunch_break_in_browser,
    lunch_break_in_ua.browser_version as lunch_break_in_browser_version,
    lunch_break_in_ua.os as lunch_break_in_os,
    lunch_break_in_ua.device as lunch_break_in_device,
    lunch_break_out_ua.browser as lunch_break_out_browser,
    lunch_break_out_ua.browser_version as lunch_break_out_browser_version,
    lunch_break_out_ua.os as lunch_break_out_os,
    lunch_break_out_ua.device as lunch_break_out_device
FROM tbl_staff_attendances sa
INNER JOIN tbl_staffs s ON s.id = sa.staff_id
LEFT JOIN tbl_useragents in_ua ON in_ua.id = sa.in_useragent_id
LEFT JOIN tbl_useragents out_ua ON out_ua.id = sa.out_useragent_id
LEFT JOIN tbl_useragents lunch_break_in_ua ON lunch_break_in_ua.id = sa.lunch_break_in_useragent_id
LEFT JOIN tbl_useragents lunch_break_out_ua ON lunch_break_out_ua.id = sa.lunch_break_out_useragent_id
WHERE
    sa.for_date = ?
    AND s.deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY s.last_name ASC, s.first_name ASC;

-- name: GetStaffAttendanceByDateRange :many
SELECT
    sa.id,
    sa.staff_id,
    sa.for_date,
    sa.time_in,
    sa.time_out,
    sa.in_location,
    sa.out_location,
    sa.in_useragent_id,
    sa.out_useragent_id,
    sa.lunch_break_in,
    sa.lunch_break_out,
    sa.lunch_break_in_location,
    sa.lunch_break_out_location,
    sa.lunch_break_in_useragent_id,
    sa.lunch_break_out_useragent_id,
    sa.created_at,
    sa.updated_at,
    s.first_name,
    s.middle_name,
    s.last_name,
    in_ua.browser as in_browser,
    in_ua.browser_version as in_browser_version,
    in_ua.os as in_os,
    in_ua.device as in_device,
    out_ua.browser as out_browser,
    out_ua.browser_version as out_browser_version,
    out_ua.os as out_os,
    out_ua.device as out_device,
    lunch_break_in_ua.browser as lunch_break_in_browser,
    lunch_break_in_ua.browser_version as lunch_break_in_browser_version,
    lunch_break_in_ua.os as lunch_break_in_os,
    lunch_break_in_ua.device as lunch_break_in_device,
    lunch_break_out_ua.browser as lunch_break_out_browser,
    lunch_break_out_ua.browser_version as lunch_break_out_browser_version,
    lunch_break_out_ua.os as lunch_break_out_os,
    lunch_break_out_ua.device as lunch_break_out_device
FROM tbl_staff_attendances sa
INNER JOIN tbl_staffs s ON s.id = sa.staff_id
LEFT JOIN tbl_useragents in_ua ON in_ua.id = sa.in_useragent_id
LEFT JOIN tbl_useragents out_ua ON out_ua.id = sa.out_useragent_id
LEFT JOIN tbl_useragents lunch_break_in_ua ON lunch_break_in_ua.id = sa.lunch_break_in_useragent_id
LEFT JOIN tbl_useragents lunch_break_out_ua ON lunch_break_out_ua.id = sa.lunch_break_out_useragent_id
WHERE
    sa.for_date >= sqlc.arg('start_date')
    AND sa.for_date <= sqlc.arg('end_date')
ORDER BY sa.for_date ASC, s.last_name ASC, s.first_name ASC;

-- name: GetStaffAttendanceByDateRangeAndStaffID :many
SELECT
    sa.id,
    sa.staff_id,
    sa.for_date,
    sa.time_in,
    sa.time_out,
    sa.in_location,
    sa.out_location,
    sa.in_useragent_id,
    sa.out_useragent_id,
    sa.lunch_break_in,
    sa.lunch_break_out,
    sa.lunch_break_in_location,
    sa.lunch_break_out_location,
    sa.lunch_break_in_useragent_id,
    sa.lunch_break_out_useragent_id,
    sa.created_at,
    sa.updated_at,
    s.first_name,
    s.middle_name,
    s.last_name,
    in_ua.browser as in_browser,
    in_ua.browser_version as in_browser_version,
    in_ua.os as in_os,
    in_ua.device as in_device,
    out_ua.browser as out_browser,
    out_ua.browser_version as out_browser_version,
    out_ua.os as out_os,
    out_ua.device as out_device,
    lunch_break_in_ua.browser as lunch_break_in_browser,
    lunch_break_in_ua.browser_version as lunch_break_in_browser_version,
    lunch_break_in_ua.os as lunch_break_in_os,
    lunch_break_in_ua.device as lunch_break_in_device,
    lunch_break_out_ua.browser as lunch_break_out_browser,
    lunch_break_out_ua.browser_version as lunch_break_out_browser_version,
    lunch_break_out_ua.os as lunch_break_out_os,
    lunch_break_out_ua.device as lunch_break_out_device
FROM tbl_staff_attendances sa
INNER JOIN tbl_staffs s ON s.id = sa.staff_id
LEFT JOIN tbl_useragents in_ua ON in_ua.id = sa.in_useragent_id
LEFT JOIN tbl_useragents out_ua ON out_ua.id = sa.out_useragent_id
LEFT JOIN tbl_useragents lunch_break_in_ua ON lunch_break_in_ua.id = sa.lunch_break_in_useragent_id
LEFT JOIN tbl_useragents lunch_break_out_ua ON lunch_break_out_ua.id = sa.lunch_break_out_useragent_id
WHERE
    sa.staff_id = sqlc.arg('staff_id')
    AND sa.for_date >= sqlc.arg('start_date')
    AND sa.for_date <= sqlc.arg('end_date')
ORDER BY sa.for_date ASC, s.last_name ASC, s.first_name ASC;

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
    require_in_shop,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: CreateStaffAttendance :one
INSERT INTO tbl_staff_attendances (
    staff_id,
    for_date,
    time_in,
    time_out,
    in_location,
    out_location,
    in_useragent_id,
    out_useragent_id,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: UpdateStaffAttendanceTimeIn :one
UPDATE tbl_staff_attendances
SET
    time_in = ?,
    in_location = ?,
    in_useragent_id = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffAttendanceTimeOut :one
UPDATE tbl_staff_attendances
SET
    time_out = ?,
    out_location = ?,
    out_useragent_id = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffAttendanceLocation :one
UPDATE tbl_staff_attendances
SET
    out_location = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: CreateStaffTimeOff :one
INSERT INTO tbl_staff_time_offs (
    type,
    start_date,
    end_date,
    description,
    staff_id,
    useragent_id,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: GetStaffTimeOffsByStaffID :many
SELECT
    sto.id,
    sto.type,
    sto.start_date,
    sto.end_date,
    sto.description,
    sto.approved,
    sto.approved_by,
    sto.approved_at,
    sto.created_at,
    sto.updated_at,
    approver.first_name as approver_first_name,
    approver.middle_name as approver_middle_name,
    approver.last_name as approver_last_name
FROM tbl_staff_time_offs sto
LEFT JOIN tbl_staffs approver ON approver.id = sto.approved_by
WHERE sto.staff_id = ?
ORDER BY sto.created_at DESC;

-- name: GetAllStaffTimeOffs :many
SELECT
    sto.id,
    sto.type,
    sto.start_date,
    sto.end_date,
    sto.description,
    sto.approved,
    sto.approved_by,
    sto.approved_at,
    sto.created_at,
    sto.updated_at,
    sto.staff_id,
    staff.first_name as staff_first_name,
    staff.middle_name as staff_middle_name,
    staff.last_name as staff_last_name,
    approver.first_name as approver_first_name,
    approver.middle_name as approver_middle_name,
    approver.last_name as approver_last_name
FROM tbl_staff_time_offs sto
INNER JOIN tbl_staffs staff ON staff.id = sto.staff_id
LEFT JOIN tbl_staffs approver ON approver.id = sto.approved_by
ORDER BY sto.created_at DESC;

-- name: ApproveStaffTimeOff :one
UPDATE tbl_staff_time_offs
SET
    approved = true,
    approved_by = ?,
    approved_at = datetime('now'),
    updated_at = datetime('now')
WHERE id = ? AND approved = false
RETURNING id;

-- name: CancelStaffTimeOff :one
UPDATE tbl_staff_time_offs
SET
    approved = false,
    approved_by = NULL,
    approved_at = NULL,
    updated_at = datetime('now')
WHERE id = ? AND approved = true
RETURNING id;

-- name: UpdateStaffAttendanceLunchBreakIn :one
UPDATE tbl_staff_attendances
SET
    lunch_break_in = ?,
    lunch_break_in_location = ?,
    lunch_break_in_useragent_id = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffAttendanceLunchBreakOut :one
UPDATE tbl_staff_attendances
SET
    lunch_break_out = ?,
    lunch_break_out_location = ?,
    lunch_break_out_useragent_id = ?,
    updated_at = datetime('now')
WHERE
    staff_id = ?
    AND for_date = ?
RETURNING id;

-- name: UpdateStaffPassword :one
UPDATE tbl_staffs
SET
    password = ?,
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;

-- name: UpdateStaffProfile :one
UPDATE tbl_staffs
SET
    first_name = ?,
    middle_name = ?,
    last_name = ?,
    mobile_no = ?,
    birthdate = ?,
    date_hired = ?,
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;
