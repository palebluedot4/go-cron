package environment

type Environment string

const (
	Production  Environment = "production"
	Staging     Environment = "staging"
	Testing     Environment = "testing"
	Development Environment = "development"
)

var validEnvironments = map[Environment]bool{
	Production:  true,
	Staging:     true,
	Testing:     true,
	Development: true,
}

func (e Environment) String() string {
	return string(e)
}

func Is(currentEnv Environment, targetEnv Environment) bool {
	return currentEnv == targetEnv
}

func IsValid(env Environment) bool {
	_, exists := validEnvironments[env]
	return exists
}

func IsProduction(env Environment) bool {
	return Is(env, Production)
}

func IsStaging(env Environment) bool {
	return Is(env, Staging)
}

func IsTesting(env Environment) bool {
	return Is(env, Testing)
}

func IsDevelopment(env Environment) bool {
	return Is(env, Development)
}
