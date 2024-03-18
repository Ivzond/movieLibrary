package helpers

import "context"

func GetRoleFromContext(ctx context.Context) string {
	role, ok := ctx.Value("role").(string)
	if !ok {
		return ""
	}
	return role
}
