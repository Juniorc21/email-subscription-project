package subscribeemails

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"
)

func SubscriptionWorkflow(ctx workflow.Context, emailDetails EmailDetails) error {
	duration := 12 * time.Second

	logger := workflow.GetLogger(ctx)
	logger.Info("Subscription created", "EmailAddress", emailDetails.EmailAddress)

	err := workflow.SetQueryHandler(ctx, "GetDetails", func() (EmailDetails, error) {
		return emailDetails, nil
	})

	defer unsubscribeHandler(ctx, emailDetails)

	if err != nil {
		return err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		WaitForCancellation: true,
	})

	logger.Info("Sending welcome email", "EmailAddress", emailDetails.EmailAddress)
	emailDetails.SubscriptionCount++
	data := EmailDetails{
		EmailAddress:      emailDetails.EmailAddress,
		Message:           "Welcome! Looks like you've been signed up",
		IsSubscribed:      true,
		SubscriptionCount: emailDetails.SubscriptionCount,
	}

	err = workflow.ExecuteActivity(ctx, SendEmail, data).Get(ctx, nil)
	if err != nil {
		return err
	}

	for emailDetails.IsSubscribed {
		emailDetails.SubscriptionCount++
		data := EmailDetails{
			EmailAddress:      emailDetails.EmailAddress,
			Message:           "This is yet another email in the subcription workflow",
			IsSubscribed:      true,
			SubscriptionCount: emailDetails.SubscriptionCount,
		}

		err := workflow.ExecuteActivity(ctx, SendEmail, data).Get(ctx, nil)

		if err != nil {
			return err
		}

		logger.Info("Sent content email", "EmailAddress", emailDetails.EmailAddress)

		if err = workflow.Sleep(ctx, duration); err != nil {
			return err
		}
	}

	return nil
}

func unsubscribeHandler(ctx workflow.Context, emailDetails EmailDetails) {
	logger := workflow.GetLogger(ctx)
	newCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	if errors.Is(ctx.Err(), workflow.ErrCanceled) {
		data := EmailDetails{
			EmailAddress:      emailDetails.EmailAddress,
			Message:           "Your subscription has been canceled. Sorry to see you go!",
			IsSubscribed:      false,
			SubscriptionCount: emailDetails.SubscriptionCount,
		}
		// send cancellation email
		err := workflow.ExecuteActivity(newCtx, SendEmail, data).Get(newCtx, nil)
		if err != nil {
			logger.Error("Failed to send cancellation email", "Error", err)
		} else {
			// Cancellation received.
			logger.Info("Sent cancellation email", "EmailAddress", emailDetails.EmailAddress)
		}
	}
}
