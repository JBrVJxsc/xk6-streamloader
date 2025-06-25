#!/usr/bin/env python3
"""
Generate a large CSV file for testing streaming CSV functionality.
Creates a file with 10,000 rows to test memory efficiency.
"""

import csv
import random
import string

def random_string(length=10):
    """Generate a random string of specified length."""
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

def random_email():
    """Generate a random email address."""
    username = random_string(8)
    domain = random.choice(['gmail.com', 'yahoo.com', 'outlook.com', 'company.com'])
    return f"{username}@{domain}"

def random_phone():
    """Generate a random phone number."""
    return f"+1-{random.randint(100, 999)}-{random.randint(100, 999)}-{random.randint(1000, 9999)}"

def main():
    with open('large.csv', 'w', newline='', encoding='utf-8') as csvfile:
        fieldnames = ['id', 'name', 'email', 'phone', 'age', 'city', 'country', 'department', 'salary', 'active']
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        
        # Write header
        writer.writeheader()
        
        # Generate 10,000 rows
        cities = ['New York', 'Los Angeles', 'Chicago', 'Houston', 'Phoenix', 'Philadelphia', 
                 'San Antonio', 'San Diego', 'Dallas', 'San Jose', 'Austin', 'Jacksonville']
        countries = ['USA', 'Canada', 'UK', 'Germany', 'France', 'Japan', 'Australia', 'Brazil']
        departments = ['Engineering', 'Marketing', 'Sales', 'HR', 'Finance', 'Operations', 'Legal', 'IT']
        
        for i in range(1, 10001):
            row = {
                'id': i,
                'name': f"{random_string(6).title()} {random_string(8).title()}",
                'email': random_email(),
                'phone': random_phone(),
                'age': random.randint(22, 65),
                'city': random.choice(cities),
                'country': random.choice(countries),
                'department': random.choice(departments),
                'salary': random.randint(30000, 150000),
                'active': random.choice(['true', 'false'])
            }
            writer.writerow(row)

if __name__ == "__main__":
    main()
    print("Generated large.csv with 10,000 rows") 