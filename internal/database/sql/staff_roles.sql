-- name: GetStaffRolesByStaffID :many
SELECT role FROM tbl_staff_roles WHERE staff_id = ?;

-- name: CreateStaffRole :one
INSERT INTO tbl_staff_roles (staff_id, role, created_at, updated_at) VALUES (?, ?, datetime('now'), datetime('now')) RETURNING id;

-- name: DeleteStaffRole :one
DELETE FROM tbl_staff_roles WHERE staff_id = ? AND role = ? RETURNING id;
