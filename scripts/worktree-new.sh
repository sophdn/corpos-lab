#!/usr/bin/env bash
# Thin shim → the owned corpos-gitflow binary on PATH. This repo is a CONSUMER of
# the git-flow service: the private toolkit service owns and publishes the binary, installed via
# scripts/fetch-corpos-gitflow.sh (or built on a dev machine with the toolkit
# checkout). Do not edit. Chain cannibalize-git-flow-service-to-go.
exec corpos-gitflow new "$@"
