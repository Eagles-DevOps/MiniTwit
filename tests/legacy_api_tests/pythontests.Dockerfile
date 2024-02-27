FROM python:3.13.0a3-bullseye

WORKDIR /tests/
COPY 'minitwit_sim_api_test.py' './'

RUN pip install requests
RUN pip install -U pytest