package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermealvess/mailman"
	"github.com/guilhermealvess/mailman/generic"
)

type Car struct {
	Mark     string  `json:"mark"`
	Model    string  `json:"model"`
	Year     int     `json:"year"`
	Price    float64 `json:"price"`
	TopSpeed string  `json:"topSpeed"`
}

func (c *Car) Show() {
	fmt.Printf("%s %s, Ano: %d, Valor: R$ %.2f, Velocidade máx.: %s\n", c.Mark, c.Model, c.Year, c.Price, c.TopSpeed)
}

func Process(ctx context.Context, event mailman.Event) error {
	<-time.After(time.Second)

	var car Car
	if err := event.Bind(&car); err != nil {
		return err
	}

	car.Show()
	return nil
}

var cars = []Car{
	{
		Mark:     "BYD",
		Model:    "Dolphin Mini",
		Year:     2025,
		Price:    115000.00,
		TopSpeed: "150 km/h",
	},
	{
		Mark:     "BYD",
		Model:    "Seal",
		Year:     2025,
		Price:    230000.00,
		TopSpeed: "250 km/h",
	},
	{
		Mark:     "Volkswagen",
		Model:    "Polo GTI",
		Year:     2024,
		Price:    110000.00,
		TopSpeed: "180 km/h",
	},
}

func simulatePublisher(ch chan Car) {
	for _, car := range cars {
		ch <- car
	}
}

func main() {
	router, channel := generic.NewGenericRouter[Car](Process)
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			simulatePublisher(channel)
		}
	}()

	manager := mailman.New()
	manager.Register("show-cars-handler", router)
	manager.Run()
}
