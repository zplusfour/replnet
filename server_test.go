package main_test

import (
	"testing"

	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/leon332157/replish/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	"github.com/valyala/fasthttp"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Replish Server")
}

var _ = BeforeSuite(func() {
	go startFiber()
	go server.StartForwardServer(7373)
	go server.StartReverseProxy()
	time.Sleep(3 * time.Second)
})

var client = &fasthttp.Client{}

func startFiber() {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendString("haha")
	})

	app.Post("/*", func(c *fiber.Ctx) error {
		return c.SendString("haha")
	})

	go app.Listen("127.0.0.1:7373")
	fmt.Println("fiber started")
}

var _ = Describe("Replish Server Main", func() {
	Describe("TCP Forwarder", func() {
		It("should serve 10000 requests (POST & GET)", func() {
			Expect(makeRequests(10000, 8383)).To(Succeed())
		})
	})
	Describe("Reverse Proxy", func() {
		It("should serve 10000 requests (POST & GET)", func() {
			Expect(makeRequests(10000, 8484)).To(Succeed())
		})
	})
})

func makeRequests(n int, port int) error {
	url := fmt.Sprintf("http://127.0.0.1:%v", port)
	for x := 0; x < n; x++ {
		statusCode, _, err := client.GetTimeout(nil, url, 1000*time.Millisecond)
		if err != nil {
			return fmt.Errorf("Failed on attempt %v err: %v", x, err)
		}
		if statusCode != fasthttp.StatusOK {
			return fmt.Errorf("Unexpected status code: %d. Expecting %d", statusCode, fasthttp.StatusOK)
		}
		// Assuming GET didn't fail, POST shouldn't fail either.
		// THERE'S NO TIMEOUT FOR POST???
		statusCode, _, err = client.Post(nil, url, nil)
		if err != nil {
			return fmt.Errorf("Failed on attempt %v err: %v", x, err)
		}
		if statusCode != fasthttp.StatusOK {
			return fmt.Errorf("Unexpected status code: %d. Expecting %d", statusCode, fasthttp.StatusOK)
		}
	}
	return nil
}
