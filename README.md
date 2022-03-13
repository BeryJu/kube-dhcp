# kube-dhcp

This is a DHCP server that uses Kubernetes CRDs to store its configuration (Scopes, Leases and OptionSets).

!!! this project is in really early development. It might eat your cat.

## Requirements

The server needs access to a kubernetes cluster (via kubeconfig, environment variables or in-cluster config).

See the files in `config/samples/` on how to configure it.

## Why

Basically, I was fed up with Active Directory's DHCP Server, I didn't feel like implementing a completely new API (like Kea for example offers) into my workflows and I wanted to build something with kubebuilder.

## Demo

https://cdn.discordapp.com/attachments/360678601227763712/951611257792921670/Screen_Recording_2022-03-10_at_23.42.22.mov
