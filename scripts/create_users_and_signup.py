#!/usr/bin/env python3
"""
Script to create 50 test users and sign them up for a specific game.
Usage: python create_users_and_signup.py [--base-url BASE_URL] [--game-id GAME_ID]
"""

import requests
import argparse
import sys
from typing import Optional


def login_user(base_url: str, email: str, password: str, index: int) -> Optional[str]:
    """Login an existing user and return their auth token."""
    url = f"{base_url}/v1/auth/login"

    data = {
        "email": email,
        "password": password
    }

    headers = {
        "Content-Type": "application/json",
        "User-Agent": "VolleyMobile/1.0"  # To get token in response
    }

    try:
        response = requests.post(url, json=data, headers=headers)

        if response.status_code == 200:
            token = response.json().get("token")
            print(f"✓ Logged in existing user {index}: {email}")
            return token
        else:
            print(f"✗ Failed to login user {index}: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"✗ Error logging in user {index}: {e}")
        return None


def create_user(base_url: str, index: int) -> Optional[str]:
    """Create a user and return their auth token. If user exists, login instead."""
    url = f"{base_url}/v1/auth/register"

    email = f"testuser{index:03d}@volley.test"
    password = "TestPassword123!"

    data = {
        "firstName": f"Test",
        "lastName": f"User{index:03d}",
        "email": email,
        "password": password
    }

    headers = {
        "Content-Type": "application/json",
        "User-Agent": "VolleyMobile/1.0"  # To get token in response
    }

    try:
        response = requests.post(url, json=data, headers=headers)

        if response.status_code in [200, 201]:
            token = response.json().get("token")
            print(f"✓ Created user {index}: {data['email']}")
            return token
        elif response.status_code == 400 and "already exists" in response.text:
            # User already exists, try to login
            print(f"  User {index} already exists, attempting login...")
            return login_user(base_url, email, password, index)
        else:
            print(f"✗ Failed to create user {index}: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"✗ Error creating user {index}: {e}")
        return None


def join_game(base_url: str, game_id: str, token: str, user_index: int) -> bool:
    """Join a game with the given auth token."""
    url = f"{base_url}/v1/games/{game_id}/participation"

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json",
        "X-Client-Type": "mobile"
    }

    try:
        response = requests.post(url, headers=headers)

        if response.status_code == 200:
            participant = response.json()
            status = participant.get("status", "unknown")
            print(f"  → User {user_index} joined game with status: {status}")
            return True
        else:
            print(f"  ✗ Failed to join game for user {user_index}: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"  ✗ Error joining game for user {user_index}: {e}")
        return False


def main():
    parser = argparse.ArgumentParser(description="Create test users and sign them up for a game")
    parser.add_argument(
        "--base-url",
        default="http://localhost:8080",
        help="Base URL of the API (default: http://localhost:8080)"
    )
    parser.add_argument(
        "--game-id",
        default="b1f21b27-8686-4952-83ce-634b098318c8",
        help="Game ID to join (default: b1f21b27-8686-4952-83ce-634b098318c8)"
    )
    parser.add_argument(
        "--count",
        type=int,
        default=50,
        help="Number of users to create (default: 50)"
    )

    args = parser.parse_args()

    print(f"Creating {args.count} users and signing them up for game {args.game_id}...")
    print(f"API Base URL: {args.base_url}\n")

    successful_signups = 0
    failed_signups = 0

    for i in range(1, args.count + 1):
        # Create user and get token
        token = create_user(args.base_url, i)

        if token:
            # Join game
            if join_game(args.base_url, args.game_id, token, i):
                successful_signups += 1
            else:
                failed_signups += 1
        else:
            failed_signups += 1

        # Add a small separator every 10 users for readability
        if i % 10 == 0:
            print()

    print("\n" + "="*60)
    print(f"Summary:")
    print(f"  Total users attempted: {args.count}")
    print(f"  Successful signups: {successful_signups}")
    print(f"  Failed signups: {failed_signups}")
    print("="*60)

    return 0 if failed_signups == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
