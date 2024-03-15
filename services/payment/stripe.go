package payment

import (
	"basedpocket/extension"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

func handleStripeWebhook(app core.App, ctx echo.Context, env *extension.Env) error {
	// ==================================================================
	// The signature check is pulled directly from Stripe and it's not tested
	req := ctx.Request()
	res := ctx.Response()

	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(res.Writer, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		eventID := sentry.CaptureException(err)
		ctx.String(http.StatusServiceUnavailable, fmt.Errorf("problem with request. eventID: %s", *eventID).Error())
		return err
	}
	endpointSecret := env.STRIPE_WEBHOOK_KEY
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		eventID := sentry.CaptureException(err)
		ctx.String(http.StatusBadRequest, fmt.Errorf("error verifying webhook signature. eventID: %s", *eventID).Error())
		return err
	}
	// ==================================================================

	if err := handleStripeEvents(app, event); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return err
	}

	res.Writer.WriteHeader(http.StatusOK)
	return nil
}

func handleStripeEvents(app core.App, event stripe.Event) error {

	if event.Type == "customer.created" {
		return handleCustomerCreatedEvent(app, event)
	}
	if event.Type == "customer.deleted" {
		return handleCustomerDeletedEvent(app, event)
	}
	if event.Type == "customer.subscription.created" {
		return handleSubscriptionCreatedEvent(app, event)
	}
	if event.Type == "customer.subscription.updated" {
		return handleSubscriptionUpdatedEvent(app, event)
	}
	if event.Type == "customer.subscription.deleted" {
		return handleSubscriptionDeletedEvent(app, event)
	}

	err := fmt.Errorf("unhandled stripe event type: %s\n", event.Type)
	eventID := sentry.CaptureException(err)
	return fmt.Errorf("unhandled stripe event type. eventID: %s", *eventID)
}

// ===============================================================================

func handleCustomerCreatedEvent(app core.App, event stripe.Event) error {
	stripeCustomer, err := getStripeCustomerFromObj(event.Data.Object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	user, err := app.Dao().FindFirstRecordByData("users", "email", stripeCustomer.Email)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customers, err := app.Dao().FindCollectionByNameOrId("customers")
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	customer := models.NewRecord(customers)
	customer.Set("user", user.Id)
	customer.Set("stripe_customer_id", stripeCustomer.ID)
	customer.Set("stripe_subscription_id", nil)
	customer.Set("tier", 0)
	if err := app.Dao().SaveRecord(customer); err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return nil
}

// ===============================================================================

func handleCustomerDeletedEvent(app core.App, event stripe.Event) error {
	stripeCustomer, err := getStripeCustomerFromObj(event.Data.Object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer, err := app.Dao().FindFirstRecordByData("customers", "stripe_customer_id", stripeCustomer.ID)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	if err := app.Dao().DeleteRecord(customer); err != nil {
		return err
	}
	return nil
}

// ===============================================================================

func handleSubscriptionCreatedEvent(app core.App, event stripe.Event) error {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	tier, err := getSubscriptionTier(stripeSubscription)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer, err := app.Dao().FindFirstRecordByData("customers", "stripe_customer_id", stripeSubscription.Customer.ID)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer.Set("stripe_subscription_id", stripeSubscription.ID)
	customer.Set("tier", tier)
	if err := app.Dao().SaveRecord(customer); err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return nil
}

// ===============================================================================

func handleSubscriptionUpdatedEvent(app core.App, event stripe.Event) error {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	tier, err := getSubscriptionTier(stripeSubscription)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer, err := app.Dao().FindFirstRecordByData("customers", "stripe_subscription_id", stripeSubscription.ID)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer.Set("tier", tier)
	if err := app.Dao().SaveRecord(customer); err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return nil
}

// ===============================================================================

func handleSubscriptionDeletedEvent(app core.App, event stripe.Event) error {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer, err := app.Dao().FindFirstRecordByData("customers", "stripe_customer_id", stripeSubscription.Customer.ID)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}

	customer.Set("stripe_subscription_id", nil)
	customer.Set("tier", 0)
	if err := app.Dao().SaveRecord(customer); err != nil {
		eventID := sentry.CaptureException(err)
		return fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return nil
}

// ===============================================================================
// ===============================================================================
// ===============================================================================

func getStripeCustomerFromObj(object map[string]interface{}) (*stripe.Customer, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	var stripeCustomer *stripe.Customer
	err = json.Unmarshal(jsonCustomer, &stripeCustomer)
	if stripeCustomer == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return stripeCustomer, nil
}

func getStripeCheckoutSessionFromObj(object map[string]interface{}) (*stripe.CheckoutSession, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	var stripeStruct *stripe.CheckoutSession
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return stripeStruct, nil
}

func getStripeSubscriptionFromObj(object map[string]interface{}) (*stripe.Subscription, error) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	var stripeStruct *stripe.Subscription
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, fmt.Errorf("error handling stripe event. eventID: %s", *eventID)
	}
	return stripeStruct, nil
}

func getSubscriptionTier(subsc *stripe.Subscription) (int, error) {
	if subsc == nil {
		return 0, nil
	}
	subscTierStr := subsc.Items.Data[0].Price.Metadata["tier"]
	subscTierInt, errTier := strconv.Atoi(subscTierStr)
	if errTier != nil {
		eventID := sentry.CaptureException(errTier)
		return 0, fmt.Errorf("failed to convert tier (%v)", eventID)
	}
	return subscTierInt, nil
}