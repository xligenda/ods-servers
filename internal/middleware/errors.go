package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xligenda/ods-servers/pkg/apierrors"
)

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	if apiErr, ok := err.(*apierrors.APIError); ok {
		return ctx.Status(apiErr.StatusCode).JSON(apiErr)
	}
	return ctx.Status(apierrors.ErrInternal.StatusCode).JSON(apierrors.ErrInternal.Message)
}
