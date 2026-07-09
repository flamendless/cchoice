package forms

type AdminSuperuserStaffsListQuery struct {
	Search string `form:"search"`
}

type AdminSuperuserStaffRolesQuery struct {
	StaffID string `form:"staff_id" validate:"required"`
}

type AdminSuperuserStaffRolePath struct {
	ID string `param:"id" validate:"required"`
}

type AdminSuperuserStaffRoleActionQuery struct {
	Action string `form:"action" validate:"required,oneof=ADD REMOVE"`
}

type AdminSuperuserStaffRoleForm struct {
	Role string `form:"role" validate:"required"`
}

type AdminSuperuserStaffCreateForm struct {
	FirstName       string `form:"first_name"`
	MiddleName      string `form:"middle_name"`
	LastName        string `form:"last_name"`
	Birthdate       string `form:"birthdate"`
	Sex             string `form:"sex"`
	DateHired       string `form:"date_hired"`
	Position        string `form:"position"`
	UserType        string `form:"user_type" validate:"required,user_type"`
	Email           string `form:"email"`
	MobileNo        string `form:"mobile_no"`
	TimeInSchedule  string `form:"time_in_schedule"`
	TimeOutSchedule string `form:"time_out_schedule"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	RequireInShop   string `form:"require_in_shop"`
	Status          string `form:"status"`
}

type AdminSuperuserStaffUpdatePath struct {
	ID string `param:"id" validate:"required"`
}

type AdminSuperuserStaffUpdateForm struct {
	Status          string `form:"status" validate:"required"`
	Position        string `form:"position"`
	TimeInSchedule  string `form:"time_in_schedule"`
	TimeOutSchedule string `form:"time_out_schedule"`
	RequireInShop   string `form:"require_in_shop"`
}
