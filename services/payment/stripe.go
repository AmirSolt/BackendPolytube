package payment

import (
	"basedpocket/cmodels"
	"basedpocket/utils"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stripe/stripe-go/v76"
)

func onStripeEvents(app core.App, event stripe.Event) *utils.CError {

	if event.Type == "customer.created" {
		return onCustomerCreatedEvent(app, event)
	}
	if event.Type == "customer.deleted" {
		return onCustomerDeletedEvent(app, event)
	}
	if event.Type == "customer.subscription.created" {
		return onSubscriptionCreatedEvent(app, event)
	}
	if event.Type == "customer.subscription.updated" {
		return onSubscriptionUpdatedEvent(app, event)
	}
	if event.Type == "customer.subscription.deleted" {
		return onSubscriptionDeletedEvent(app, event)
	}

	err := fmt.Errorf("unhandled stripe event type: %s\n", event.Type)
	eventID := sentry.CaptureException(err)
	return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
}

// ===============================================================================

func onCustomerCreatedEvent(app core.App, event stripe.Event) *utils.CError {
	stripeCustomer, err := getStripeCustomerFromObj(event.Data.Object)
	if err != nil {
		return err
	}

	var user *cmodels.User
	if err := user.FindUser(app, &cmodels.FindUserParams{Email: stripeCustomer.Email}); err != nil {
		return err
	}

	customer := &cmodels.Customer{
		User:                 user.Id,
		StripeCustomerID:     stripeCustomer.ID,
		StripeSubscriptionID: "",
		Tier:                 0,
	}
	if err := customer.SaveCustomer(app); err != nil {
		return err
	}
	return nil
}

// ===============================================================================

func onCustomerDeletedEvent(app core.App, event stripe.Event) *utils.CError {
	stripeCustomer, err := getStripeCustomerFromObj(event.Data.Object)
	if err != nil {
		return err
	}

	var customer *cmodels.Customer
	if err := customer.FindCustomer(app, &cmodels.FindCustomerParams{StripeCustomerID: stripeCustomer.ID}); err != nil {
		return err
	}

	if err := customer.DeleteCustomer(app); err != nil {
		return err
	}
	return nil
}

// ===============================================================================

func onSubscriptionCreatedEvent(app core.App, event stripe.Event) *utils.CError {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		return err
	}

	tier, err := getSubscriptionTier(stripeSubscription)
	if err != nil {
		return err
	}

	var customer *cmodels.Customer
	if err := customer.FindCustomer(app, &cmodels.FindCustomerParams{StripeCustomerID: stripeSubscription.Customer.ID}); err != nil {
		return err
	}
	customer.StripeSubscriptionID = stripeSubscription.ID
	customer.Tier = tier
	if err := customer.SaveCustomer(app); err != nil {
		return err
	}
	return nil
}

// ===============================================================================

func onSubscriptionUpdatedEvent(app core.App, event stripe.Event) *utils.CError {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		return err
	}

	tier, err := getSubscriptionTier(stripeSubscription)
	if err != nil {
		return err
	}

	var customer *cmodels.Customer
	if err := customer.FindCustomer(app, &cmodels.FindCustomerParams{StripeCustomerID: stripeSubscription.Customer.ID}); err != nil {
		return err
	}
	customer.Tier = tier
	if err := customer.SaveCustomer(app); err != nil {
		return err
	}
	return nil
}

// ===============================================================================

func onSubscriptionDeletedEvent(app core.App, event stripe.Event) *utils.CError {
	stripeSubscription, err := getStripeSubscriptionFromObj(event.Data.Object)
	if err != nil {
		return err
	}

	var customer *cmodels.Customer
	if err := customer.FindCustomer(app, &cmodels.FindCustomerParams{StripeCustomerID: stripeSubscription.Customer.ID}); err != nil {
		return err
	}
	if stripeSubscription.ID == customer.StripeSubscriptionID {
		customer.StripeSubscriptionID = ""
		customer.Tier = 0
		if err := customer.SaveCustomer(app); err != nil {
			return err
		}
	}
	return nil
}

// ===============================================================================
// ===============================================================================
// ===============================================================================

func getStripeCustomerFromObj(object map[string]interface{}) (*stripe.Customer, *utils.CError) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	var stripeCustomer *stripe.Customer
	err = json.Unmarshal(jsonCustomer, &stripeCustomer)
	if stripeCustomer == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return stripeCustomer, nil
}

func getStripeCheckoutSessionFromObj(object map[string]interface{}) (*stripe.CheckoutSession, *utils.CError) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	var stripeStruct *stripe.CheckoutSession
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return stripeStruct, nil
}

func getStripeSubscriptionFromObj(object map[string]interface{}) (*stripe.Subscription, *utils.CError) {
	jsonCustomer, err := json.Marshal(object)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	var stripeStruct *stripe.Subscription
	err = json.Unmarshal(jsonCustomer, &stripeStruct)
	if stripeStruct == nil || err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return stripeStruct, nil
}

func getSubscriptionTier(subsc *stripe.Subscription) (int, *utils.CError) {
	if subsc == nil {
		return 0, nil
	}
	subscTierStr := subsc.Items.Data[0].Price.Metadata["tier"]
	subscTierInt, err := strconv.Atoi(subscTierStr)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return 0, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return subscTierInt, nil
}
