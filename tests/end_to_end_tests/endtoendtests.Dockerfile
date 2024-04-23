FROM python:3.11-slim-bullseye

WORKDIR /tests/
COPY 'test_itu_minitwit_ui.py' './'

RUN apt-get update && apt-get install -y \
    libpq-dev \
    gcc \
    wget \ 
    gnupg \
    unzip \
    && rm -rf /var/lib/apt/lists/*

RUN pip install selenium
RUN pip install psycopg2-binary
RUN pip install pytest
RUN pip install webdriver-manager


RUN wget https://chromedriver.storage.googleapis.com/94.0.4606.41/chromedriver_linux64.zip \
    && unzip chromedriver_linux64.zip \
    && mv chromedriver /usr/local/bin/ \
    && chmod +x /usr/local/bin/chromedriver \
    && rm chromedriver_linux64.zip