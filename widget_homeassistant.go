package main

import (
	"context"
	"fmt"
	ha "github.com/mkelcik/go-ha-client"
	"net/http"
	"strings"
	"time"
)

type HomeAssistantWidget struct {
	*ButtonWidget

	entityId           string
	client             *ha.Client
	iconOn             string
	iconOff            string
	iconNoPlayer       string
	entityName         bool
	token              string
	url                string
	currentArtURL      string
	service            string
	currentEntityState string
}

func NewHomeAssistantWidget(bw *BaseWidget, opts WidgetConfig) (*HomeAssistantWidget, error) {
	widget, err := NewButtonWidget(bw, opts)
	if err != nil {
		return nil, err
	}

	bw.setInterval(time.Duration(opts.Interval)*time.Millisecond, 500)

	var entityId, iconOn, iconOff, url, token, service string
	var entityName bool
	_ = ConfigValue(opts.Config["entity_id"], &entityId)
	_ = ConfigValue(opts.Config["icon_on"], &iconOn)
	_ = ConfigValue(opts.Config["icon_off"], &iconOff)
	_ = ConfigValue(opts.Config["url"], &url)
	_ = ConfigValue(opts.Config["token"], &token)
	_ = ConfigValue(opts.Config["entity_name"], &entityName)
	_ = ConfigValue(opts.Config["service"], &service)

	client := ha.NewClient(ha.ClientConfig{
		Token: token,
		Host:  url,
	}, &http.Client{
		Timeout: 30 * time.Second,
	})

	// Ping the Home Assistant instance
	if err := client.Ping(context.Background()); err != nil {
		fmt.Println("HomeAssistant connection error", err)
		return nil, err
	}

	return &HomeAssistantWidget{
		ButtonWidget: widget,
		entityId:     entityId,
		iconOn:       iconOn,
		iconOff:      iconOff,
		entityName:   entityName,
		client:       client,
		service:      service,
	}, nil
}

func (w *HomeAssistantWidget) Update() error {
	fresh := true
	stateEntity, err := w.client.GetStateForEntity(context.Background(), w.entityId)

	if err != nil {
		fmt.Println(err)
	}
	if w.entityName {
		w.label = ""
		if stateEntity.Attributes["friendly_name"] != nil {
			w.label = stateEntity.Attributes["friendly_name"].(string)
		}

	}

	status := stateEntity.State
	if status != w.currentEntityState {
		w.currentEntityState = status
		fresh = false

		if w.currentEntityState == "on" {
			if err := w.LoadImage(w.iconOn); err != nil {
				return err
			}
		} else {
			if err := w.LoadImage(w.iconOff); err != nil {
				return err
			}
		}
	}

	if !fresh {
		return w.ButtonWidget.Update()
	}

	return nil
}

func (w *HomeAssistantWidget) TriggerAction(hold bool) {
	parts := strings.Split(w.service, ".")
	if len(parts) > 0 {
		domain := parts[0]
		service := parts[1]
		if _, err := w.client.CallService(context.Background(), ha.DefaultServiceCmd{
			Service:  service,
			Domain:   domain,
			EntityId: w.entityId,
		}); err != nil {
			panic(err)
		}
	}
}
