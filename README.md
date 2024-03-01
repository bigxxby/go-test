Golang and PostgreSQL Application README

This repository contains a simple Golang application that interacts with a PostgreSQL database using Docker Compose. The application provides functionality to manage data related to 'goods' through a RESTful API.
Prerequisites

    Docker
    Docker Compose

Installation and Setup

    Clone this repository to your local machine.
    Make sure you have Docker and Docker Compose installed and running.
    Navigate to the project directory.
    Run docker-compose up --build to build and start the application containers.

Routes
GET /good/get

    Description: Retrieves information about goods from the database.
    Method: GET
    Response: JSON format containing information about goods.

POST /good/create

    Description: Creates a new entry for a good in the database.
    Method: POST
    Request Body: JSON format representing the details of the good to be created.
    Response: JSON format containing the created good's details.

POST /good/update

    Description: Updates an existing entry for a good in the database.
    Method: PATCH
    Request Body: JSON format representing the updated details of the good.
    Response: JSON format containing the updated good's details.

POST /good/remove

    Description: Removes an existing entry for a good from the database.
    Method: DELETE
    Request Body: JSON format containing the ID of the good to be removed.
    Response: JSON format confirming the removal of the good.
