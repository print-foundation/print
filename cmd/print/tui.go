package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/print-foundation/print/internal/builder"
	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/install"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/mirrors"
	"github.com/rivo/tview"
)

func col(hex string) tcell.Color {
	return tcell.GetColor(hex)
}

func runTUI(log logging.Logger) error {
	app := tview.NewApplication()
	state := struct {
		distro  mirrors.Distro
		country string
		mirror  mirrors.Mirror
		ssid    string
		pass    string
		arch    builder.Arch
		out     string
		device  string
		flash   bool
	}{arch: builder.Arch(runtime.GOARCH)}

	bg, fg := tcell.GetColor("#0d1117"), tcell.GetColor("#e6edf3")

	var (
		distroList *tview.List
		mirrorList *tview.List
		progress   *tview.TextView
	)

	var showDistro, showCountry, showMirror, showWiFi, showProgress func()

	showDistro = func() {
		distroList = tview.NewList().ShowSecondaryText(false)
		distroList.SetBorder(true).SetTitle(" 1. Choose a distro ")
		distroList.SetBackgroundColor(col("#0d1117"))
		for _, d := range mirrors.ListDistros() {
			d := d
			distroList.AddItem(string(d), "", 0, func() {
				state.distro = d
				showCountry()
			})
		}
		app.SetRoot(distroList, true)
	}

	showCountry = func() {
		form := tview.NewForm()
		form.SetBorder(true).SetTitle(" 2. Country (ISO-3166, e.g. US, DE, FR) ")
		form.SetBackgroundColor(bg)
		form.SetFieldTextColor(fg)
		form.AddInputField("Country", "US", 2, func(text string, _ rune) bool { return len(text) < 3 },
			func(t string) { state.country = strings.ToUpper(strings.TrimSpace(t)) })
		form.AddButton("Next", func() {
			if state.country == "" {
				state.country = "US"
			}
			showMirror()
		})
		form.AddButton("Back", showDistro)
		app.SetRoot(form, true)
	}

	showMirror = func() {
		resolver := mirrors.NewResolver()
		ms, err := resolver.Resolve(context.Background(), state.distro, state.country)
		if err != nil || len(ms) == 0 {
			modal := tview.NewModal().SetText("No mirror found for " + state.country + ". Try another country.").
				AddButtons([]string{"Back"}).SetDoneFunc(func(int, string) { showCountry() })
			app.SetRoot(modal, true)
			return
		}
		mirrorList = tview.NewList().ShowSecondaryText(true)
		mirrorList.SetBorder(true).SetTitle(" 3. Choose a mirror ")
		mirrorList.SetBackgroundColor(bg)
		for _, m := range ms {
			m := m
			mirrorList.AddItem(m.Host, m.BaseURL, 0, func() {
				state.mirror = m
				showWiFi()
			})
		}
		app.SetRoot(mirrorList, true)
	}

	showWiFi = func() {
		form := tview.NewForm()
		form.SetBorder(true).SetTitle(" 4. WiFi (optional - leave blank for wired) ")
		form.SetBackgroundColor(bg)
		form.SetFieldTextColor(fg)
		form.AddInputField("Architecture (amd64/arm64)", string(state.arch), 8, nil, func(t string) {
			a := strings.ToLower(strings.TrimSpace(t))
			if a == "amd64" || a == "arm64" {
				state.arch = builder.Arch(a)
			}
		})
		form.AddInputField("SSID", "", 40, nil, func(t string) { state.ssid = t })
		form.AddPasswordField("Passphrase", "", 40, '*', func(t string) { state.pass = t })
		form.AddInputField("Output ISO path", "print-"+string(state.distro)+".iso", 40, nil, func(t string) { state.out = t })
		form.AddInputField("Flash device (e.g. /dev/sdb, blank=skip)", "", 40, nil, func(t string) { state.device = t })
		form.AddButton("Build", func() {
			if state.out == "" {
				state.out = "print-" + string(state.distro) + ".iso"
			}
			dev := strings.TrimSpace(state.device)
			if dev != "" {
				modal := tview.NewModal().
					SetText(fmt.Sprintf("You are about to write the ISO to %s. This will ERASE all data on that device.\nProceed?", dev)).
					AddButtons([]string{"Write", "Cancel"}).
					SetDoneFunc(func(_ int, label string) {
						if label == "Write" {
							state.flash = true
							showProgress()
						} else {
							state.flash = false
							showWiFi()
						}
					})
				app.SetRoot(modal, true)
				return
			}
			state.flash = false
			showProgress()
		})
		form.AddButton("Back", showMirror)
		app.SetRoot(form, true)
	}

	showProgress = func() {
		progress = tview.NewTextView().SetDynamicColors(true)
		progress.SetBorder(true).SetTitle(" Building " + string(state.distro) + " ISO ")
		progress.SetBackgroundColor(bg)
		app.SetRoot(progress, true)
		go func() {
			clientBin, err := buildClient(log, state.arch)
			if err != nil {
				app.QueueUpdateDraw(func() {
					progress.SetText("[red]" + err.Error() + "[-]")
				})
				return
			}
			spec := builder.Spec{
				Distro: state.distro,
				Arch:   state.arch,
				Mirror: state.mirror,
				Output: state.out,
				WiFi:   builder.WiFiConfig{SSID: state.ssid, Passphrase: state.pass},
			}
			result, err := builder.New(builder.WithLogger(log)).Build(context.Background(), spec, clientBin)
			msg := ""
			if err != nil {
				msg = "[red]" + err.Error() + "[-]"
			} else {
				msg = "[green]built " + result.ISO + " (verified=" + fmt.Sprintf("%v", result.Verified) + ")[-]\n"
				if state.flash && state.device != "" {
					eng := install.NewEngine(nil, nil, install.WithLogger(log), install.WithDeviceWriter(install.FileDeviceWriter{}))
					ferr := eng.FlashLocal(context.Background(), result.ISO, state.device, disk.Confirmation{DevicePath: state.device, Acknowledged: true}, nil)
					if ferr != nil {
						msg += "[red]flash failed: " + ferr.Error() + "[-]\n"
					} else {
						msg += "[green]flashed to " + state.device + "[-]\n"
					}
				}
			}
			app.QueueUpdateDraw(func() {
				progress.SetText(msg)
			})
		}()
	}

	showDistro()
	return app.Run()
}
