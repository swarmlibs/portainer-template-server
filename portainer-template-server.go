package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
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

func fetchRepos(reposURL string) ([]string, error) {
	if reposURL != "" {
		resp, err := http.Get(reposURL)
		if err != nil {
			log.Fatalf("Failed to fetch template list from %s: %v\n", reposURL, err)
		}
		defer resp.Body.Close()
		respTemplateURLs := []string{}
		if err := json.NewDecoder(resp.Body).Decode(&respTemplateURLs); err != nil {
			log.Fatalf("Failed to decode template list from %s: %v\n", reposURL, err)
		}
		return respTemplateURLs, nil
	}
	return make([]string, 0), nil
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
			&cli.StringFlag{
				Name:  "repos-url",
				Usage: "URL to a list of template URLs",
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

			reposURL := c.String("repos-url")
			templateURLs := c.StringSlice("template-url")

			if reposURL != "" {
				if respTemplateURLs, err := fetchRepos(reposURL); err == nil {
					log.Printf("Repositories URL: %s\n", reposURL)
					for _, url := range respTemplateURLs {
						if slices.Contains(templateURLs, url) {
							continue
						}
						templateURLs = append(templateURLs, url)
					}
				}
			}

			for _, url := range templateURLs {
				log.Printf("Serving template from %s\n", url)
			}

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				list := []string{}
				for _, url := range templateURLs {
					list = append(list, url)
				}
				if err := json.NewEncoder(w).Encode(list); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})
			mux.HandleFunc("/templates.json", func(w http.ResponseWriter, r *http.Request) {
				combinedAppTemplateScheme := PortainerAppTemplateScheme{
					Version:   c.String("template-version"),
					Templates: []map[string]any{},
				}
				for _, url := range templateURLs {
					appTemplate := PortainerAppTemplate{
						Url: url,
					}
					if err := appTemplate.FetchTemplate(); err != nil {
						log.Printf("Failed to fetch template from %s: %v\n", url, err)
						continue
					}
					combinedAppTemplateScheme.Templates = append(combinedAppTemplateScheme.Templates, appTemplate.Scheme.Templates...)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(combinedAppTemplateScheme); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			})
			mux.HandleFunc("/-/health", (func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			}))
			mux.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
				if reposURL != "" {
					if respTemplateURLs, err := fetchRepos(reposURL); err == nil {
						for _, url := range respTemplateURLs {
							if slices.Contains(templateURLs, url) {
								continue
							}
							templateURLs = append(templateURLs, url)
						}
					}
					log.Print("Reloaded templates")
				}
				w.Write([]byte("OK"))
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
