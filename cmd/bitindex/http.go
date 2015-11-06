package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const StatusUnprocessableEntity = 422

type query struct {
	Any  []uint32
	Nany []uint32
	All  []uint32
	Nall []uint32
}

var httpCmd = &cobra.Command{
	Use: "http <index>",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Println("An index file is required.")
			os.Exit(1)
		}

		f, err := os.Open(args[0])

		if err != nil {
			cmd.Println("Error opening file:", err)
			os.Exit(1)
		}

		var idx *bitindex.Index

		if idx, err = bitindex.LoadIndex(f); err != nil {
			cmd.Println("Error loading index file:", err)
			os.Exit(1)
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")

			v := map[string]interface{}{
				"domain_size":    idx.Domain.Size(),
				"table_size":     idx.Table.Size(),
				"index_sparsity": idx.Sparsity(),
			}

			if err := json.NewEncoder(w).Encode(v); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
		})

		http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")

			// Expected up to 4 keys, one for each operator.
			q := query{}

			defer r.Body.Close()

			if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
				w.WriteHeader(StatusUnprocessableEntity)
				fmt.Fprint(w, err)
				return
			}

			res, err := idx.Query(q.Any, q.All, q.Nany, q.Nall)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}

			if err = json.NewEncoder(w).Encode(res); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
		})

		http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")

			if err := json.NewEncoder(w).Encode(idx.Table.Keys()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
		})

		http.HandleFunc("/domain", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")

			if err := json.NewEncoder(w).Encode(idx.Domain.Members()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
		})

		addr := fmt.Sprintf("%s:%d", viper.GetString("http.host"), viper.GetInt("http.port"))
		cmd.Printf("Listening on %s...\n", addr)

		http.ListenAndServe(addr, nil)
	},
}

func init() {
	flags := httpCmd.Flags()

	flags.String("host", "127.0.0.1", "Host of the HTTP server.")
	flags.Int("port", 7000, "Port of the HTTP server.")

	viper.BindPFlag("http.host", flags.Lookup("host"))
	viper.BindPFlag("http.port", flags.Lookup("port"))
}
