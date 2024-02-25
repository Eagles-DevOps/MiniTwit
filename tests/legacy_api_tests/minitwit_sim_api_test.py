import os
import json
import base64
import requests
import unittest

BASE_URL = "http://localhost:15001"
USERNAME = 'simulator'
PWD = 'super_safe!'
CREDENTIALS = ':'.join([USERNAME, PWD]).encode('ascii')
ENCODED_CREDENTIALS = base64.b64encode(CREDENTIALS).decode()
HEADERS = {'Connection': 'close',
           'Content-Type': 'application/json',
           f'Authorization': f'Basic {ENCODED_CREDENTIALS}'}

class MiniTwitSimApi(unittest.TestCase):
    
    def test_aa(self):
        url = f"{BASE_URL}/delete"
        response = requests.post(url, headers=HEADERS)
        assert response.ok

    def test_a_latest(self):
        # post something to updaet LATEST
        url = f"{BASE_URL}/register"
        data = {'username': 'test', 'email': 'test@test', 'pwd': 'foo'}
        params = {'latest': 1337}
        response = requests.post(url, data=json.dumps(data),
                                params=params, headers=HEADERS)
        assert response.ok

        # verify that latest was updated
        url = f'{BASE_URL}/latest'
        response = requests.get(url, headers=HEADERS)
        assert response.ok
        assert response.json()['latest'] == 1337


    def test_b_register(self):
        username = 'a'
        email = 'a@a.a'
        pwd = 'a'
        data = {'username': username, 'email': email, 'pwd': pwd}
        params = {'latest': 1}
        response = requests.post(f'{BASE_URL}/register',
                                data=json.dumps(data), headers=HEADERS, params=params)
        assert response.ok
        # TODO: add another assertion that it is really there

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 1


    def test_c_create_msg(self):
        username = 'a'
        data = {'content': 'Blub!'}
        url = f'{BASE_URL}/msgs/{username}'
        params = {'latest': 2}
        response = requests.post(url, data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 2


    def test_d_get_latest_user_msgs(self):
        username = 'a'

        query = {'no': 20, 'latest': 3}
        url = f'{BASE_URL}/msgs/{username}'
        response = requests.get(url, headers=HEADERS, params=query)
        assert response.status_code == 200

        got_it_earlier = False
        for msg in response.json():
            if msg['content'] == 'Blub!' and msg['user'] == username:
                got_it_earlier = True

        assert got_it_earlier

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 3


    def test_e_get_latest_msgs(self):
        username = 'a'
        query = {'no': 20, 'latest': 4}
        url = f'{BASE_URL}/msgs'
        response = requests.get(url, headers=HEADERS, params=query)
        assert response.status_code == 200

        got_it_earlier = False
        for msg in response.json():
            if msg['content'] == 'Blub!' and msg['user'] == username:
                got_it_earlier = True

        assert got_it_earlier

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 4


    def test_f_register_b(self):
        username = 'b'
        email = 'b@b.b'
        pwd = 'b'
        data = {'username': username, 'email': email, 'pwd': pwd}
        params = {'latest': 5}
        response = requests.post(f'{BASE_URL}/register', data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok
        # TODO: add another assertion that it is really there

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 5


    def test_g_register_c(self):
        username = 'c'
        email = 'c@c.c'
        pwd = 'c'
        data = {'username': username, 'email': email, 'pwd': pwd}
        params = {'latest': 6}
        response = requests.post(f'{BASE_URL}/register', data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 6


    def test_h_follow_user(self):
        username = 'a'
        url = f'{BASE_URL}/fllws/{username}'
        data = {'follow': 'b'}
        params = {'latest': 7}
        response = requests.post(url, data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok

        data = {'follow': 'c'}
        params = {'latest': 8}
        response = requests.post(url, data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok

        query = {'no': 20, 'latest': 9}
        response = requests.get(url, headers=HEADERS, params=query)
        assert response.ok

        json_data = response.json()
        assert "b" in json_data["follows"]
        assert "c" in json_data["follows"]

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 9


    def test_i_a_unfollows_b(self):
        username = 'a'
        url = f'{BASE_URL}/fllws/{username}'

        #  first send unfollow command
        data = {'unfollow': 'b'}
        params = {'latest': 10}
        response = requests.post(url, data=json.dumps(data),
                                headers=HEADERS, params=params)
        assert response.ok

        # then verify that b is no longer in follows list
        query = {'no': 20, 'latest': 11}
        response = requests.get(url, params=query, headers=HEADERS)
        assert response.ok
        assert 'b' not in response.json()['follows']

        # verify that latest was updated
        response = requests.get(f'{BASE_URL}/latest', headers=HEADERS)
        assert response.json()['latest'] == 11

if __name__ == '__main__':
    unittest.main()