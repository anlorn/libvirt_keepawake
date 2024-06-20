# Libvirt Keepawake

## Description
A tool to keep the host PC awake(inhibit sleep) while a VM is running. 


## How it works

It uses "libvirt" to track running VMs. To inhibit/uninhibit sleep, it uses the DBUS method "org. free desktop.PowerManagement.Inhibit.Inhibit"(destination "org.freedesktop.PowerManagement" and path "/org/freedesktop/PowerManagement/Inhibit").

I use xfce, and this method is provided by xfce-power-manager. To test if this method provided run `dbus-send --session --print-reply --dest=org.freedesktop.PowerManagement /org/freedesktop/PowerManagement/Inhibit org.freedesktop.PowerManagement.Inhibit.Inhibit string:"YourAppName" string:"ReasonForInhibition"`

Application exists and remove on all active sleep inhibitors on SIGKILL and SIGHUP. So, it can be safely autostarted on user login.

## Installation

* Ensure libvirt is installed(see Dockerfile for dependencies)
* Optionally you can run `make test`
* Run `make build`
* Copy binary `libvirt-keepwake` to `~/.local/bin/`.
* Copy `libvirt-keepawake.desktop` from the `init` folder to `~/.config/autostart/.` This will make the application start when the user logs in. 

Gentoo users can skip first and second steps and just use ebuild from init folder.

## My Use case

I have a desktop with a powerful GPU for ML, which is my main working/fun/projects PC.
It has Gentoo(so no systemd) as OS. However, using the same GPU to play games on the PC,
laptop, or TV downstairs is convenient. So, I use libvirt + qemu to run Windows 11
with GPU(and a few other PCI devices) passthrough. I use [Sunshine](https://github.com/LizardByte/Sunshine) 
installed on Windows 11  to stream games to different devices (For example, [Moonlight](https://github.com/moonlight-stream) 
on RPI5 connected to a TV). I want the host PC to stay awake while the Windows 11 VM is active.