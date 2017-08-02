# Dux

## Introduction

Dux is a tool that allows you to forward TCP requests to a docker container using only the [docker API](https://docs.docker.com/engine/api/).

## But, why?

More and more we're seeing clusters of machines set up to which users don't have SSH access but they do have docker API access. This is the case whever you are a user of [Docker Datacenter](https://www.docker.com/enterprise-edition), or if you deploy a swarm cluster with [Docker Cloud](https://docs.docker.com/docker-cloud/cloud-swarm/), or maybe even you set up your own infrastructure and don't want to give cluster users SSH access.

One of the things I used to love about being able to SSH into dev boxes was using [SSH port forwarding](https://help.ubuntu.com/community/SSH/OpenSSH/PortForwarding). This feature let you expose a socket locally and have all the requests to that socket actually made by the remote machine. So if you had a web server running remotely you could connect to it with your local browser.

## Enter Dux

Dux gives you the same functionality as SSH port forwarding using the Docker API. So now you can debug your docker services without opening up your networks and take advantage of all the great features of DDC and Docker Cloud like access control and Destop2Cloud.

## How it works

Dux basically has two parts:

1. **dux** is what you run locally. It exposes a socket either as a file or via TCP and muxes the result into the stdin of **duxd**.
2. **duxd** runs remotely and demuxes the requests on the other end.
