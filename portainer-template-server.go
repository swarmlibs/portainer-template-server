package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2" // imports as package "cli"
)

var (
	DefaultTemplateURLs = []string{
		// "https://raw.githubusercontent.com/portainer/templates/v3/templates.json",
		// "https://raw.githubusercontent.com/swarmlibs/portainer-templates/refs/heads/main/templates.json",
	}
)

type PortainerAppTemplateScheme struct {
	Version   string           `json:"version"`
	Templates []map[string]any `json:"templates"`
}

type PortainerAppTemplate struct {
	Url    string
	Scheme PortainerAppTemplateScheme
}

func (t *PortainerAppTemplate) FetchTemplate() error {
	// fetch template from URL
	resp, err := http.Get(t.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch template")
	}

	if err := json.NewDecoder(resp.Body).Decode(&t.Scheme); err != nil {
		return err
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "portainer-template-server",
		Usage: "Portainer template server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "Host to listen on",
				Value: "0.0.0.0",
			},
			&cli.StringFlag{
				Name:  "port",
				Usage: "Port to listen on",
				Value: "4242",
			},
			&cli.StringFlag{
				Name:  "template-version",
				Usage: "Set the version of the template response",
				Value: "3",
			},
			&cli.StringSliceFlag{
				Name:  "template-url",
				Usage: "URL to a template file",
				Value: cli.NewStringSlice(DefaultTemplateURLs...),
			},
		},
		Action: func(c *cli.Context) error {
			host := c.String("host")
			port := c.String("port")

			mux := http.NewServeMux()
			server := &http.Server{
				Addr:    host + ":" + port,
				Handler: mux,
			}
			log.Printf("Starting server on %s\n", server.Addr)

			combinedAppTemplateScheme := PortainerAppTemplateScheme{
				Version:   c.String("template-version"),
				Templates: []map[string]any{},
			}

			templateURLs := c.StringSlice("template-url")
			for _, url := range templateURLs {
				appTemplate := PortainerAppTemplate{
					Url: url,
				}
				if err := appTemplate.FetchTemplate(); err != nil {
					log.Printf("Failed to fetch template from %s: %v\n", url, err)
					continue
				}
				log.Printf("Serving template from %s\n", url)
				combinedAppTemplateScheme.Templates = append(combinedAppTemplateScheme.Templates, appTemplate.Scheme.Templates...)
			}

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(combinedAppTemplateScheme); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})

			shutdownChan := make(chan bool, 1)

			go func() {
				if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					log.Fatalf("HTTP server error: %v", err)
				}
				time.Sleep(1 * time.Millisecond)
				log.Println("Stopped serving new connections.")
				shutdownChan <- true
			}()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownRelease()

			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Fatalf("HTTP shutdown error: %v", err)
			}

			<-shutdownChan
			log.Println("Graceful shutdown complete.")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
