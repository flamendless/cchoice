package enums
type AppEnv string
const (
	APP_ENV_LOCAL AppEnv = "local"
	APP_ENV_WEB   AppEnv = "web"
	APP_ENV_PROD  AppEnv = "prod"
)
func (ae AppEnv) String() string {
	return string(ae)
}
