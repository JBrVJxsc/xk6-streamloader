import json

data = [
    {
        "method": "GET",
        "requestURI": f"/bulk/{i}",
        "headers": {"X": str(i)},
        "content": str(i)
    }
    for i in range(1000)
]

with open("large.json", "w") as f:
    json.dump(data, f, indent=2)
