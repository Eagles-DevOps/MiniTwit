FROM python:3.11-slim-bullseye

WORKDIR /tests/
COPY 'test_itu_minitwit_ui.py' './'

RUN apt-get update && apt-get install -y \
    libpq-dev \
    gcc \
    wget  # Including wget here

RUN pip install selenium
RUN pip install psycopg2-binary
RUN pip install pytest
RUN wget https://github.com/mozilla/geckodriver/releases/download/v0.32.0/geckodriver-v0.32.0-linux64.tar.gz
RUN tar xzvf geckodriver-v0.32.0-linux64.tar.gz
RUN rm geckodriver-v0.32.0-linux64.tar.gz
