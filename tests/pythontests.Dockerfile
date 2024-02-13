FROM python:3.13.0a3-bullseye

WORKDIR /tests/
COPY 'refactored_tests.py' './'

RUN pip install requests