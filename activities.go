package subscribeemails

import (
	"context"

	"go.temporal.io/sdk/activity"
)

func SendEmail(ctx context.Context, emailInfo EmailDetails) (string, error) {
	activity.GetLogger(ctx).Info("Sending email to customer", "EmailAddress", emailInfo.EmailAddress)
	return "Email sent to " + emailInfo.EmailAddress, nil
}
