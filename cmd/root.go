package cmd

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	libvirtLibrary "libvirt.org/go/libvirt"
	"libvirt_keepawake/internal"
	"libvirt_keepawake/internal/libvirt_watcher"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "Starts daemon",
	Short: "Starts Daemon",
	Long:  `Start Daemon`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose, err := cmd.Flags().GetBool("verbose"); err != nil {
			log.WithError(err).Error("Can't get verbose flag value")
			return
		} else if verbose {
			log.SetLevel(log.DebugLevel)
			log.Debug("Verbose mode enabled")
		} else {
			log.SetLevel(log.InfoLevel)
		}
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		conn, err := dbus.SessionBusPrivateNoAutoStartup()
		if err != nil {
			log.WithError(err).Error("Can't connect to session DBUS")
			os.Exit(1)
		}
		err = conn.Auth(nil)
		if err != nil {
			log.WithError(err).Error("Can't authenticate to DBUS")
			os.Exit(1)
		}
		log.Debug("Successfully authenticated to DBUS")

		err = conn.Hello()
		if err != nil {
			log.WithError(err).Error("Failed to send hello to DBUS after connection")
			os.Exit(1)
		}
		log.Debug("Successfully sent hello to DBUS")

		log.Info("Successfully connected to session DBUS")

		sleepInhibitor := internal.NewDbusSleepInhibitor(conn)

		// how to listen for libvirt event
		libVirtConn, libVirtConErr := libvirtLibrary.NewConnect("qemu:///system")
		if libVirtConErr != nil {
			log.WithError(libVirtConErr).Error("Can't connect to libvirt")
			os.Exit(1)
		} else {
			log.Debug("Successfully connected to libvirt")
			defer func() {
				_, err := libVirtConn.Close()
				if err != nil {
					log.WithError(err).Error("Can't close libvirt connection")
				}
			}()
		}
		connAdapter := libvirt_watcher.LibvirtConnectAdapter{Connect: libVirtConn}
		watcher := libvirt_watcher.NewLibvirtWatcher(&connAdapter)

		ticker := time.NewTicker(10 * time.Second)

		orchestrator := internal.NewOrchestrator(sleepInhibitor, watcher, ticker)
		orchestrator.Start()
		defer func() {
			log.Debug("Stopping orchestrator")
			orchestrator.Stop()
			// TODO fix	this
			time.Sleep(5 * time.Second)
			if conn != nil {
				err := conn.Close()
				if err != nil {
					log.WithError(err).Error("Can't close DBUS connection")
				}
			}
		}()
		log.Info("Will wait for interrupt signal")
		v, ok := <-ch
		if !ok {
			fmt.Println(ok)
		}
		fmt.Println(v)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
