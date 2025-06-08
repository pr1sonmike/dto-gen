# DTO generation tool

## Description

A tool for generating Data Transfer Objects (DTOs) from Go structs.

## Prerequisites

- Go 1.24 or later installed.
- A valid `go.mod` file in your project.

## Usage

Run the following command to generate a DTO:

```bash
go run main.go --input example/model.go --output example/user_dto.go --type User
