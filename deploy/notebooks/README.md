# JupyterLab Notebooks

This directory contains Jupyter notebooks for data analysis, testing, and development.

## Getting Started

1. **Access JupyterLab:**
   ```
   http://localhost:8889
   ```

2. **Login:**
   - Password: Use `JUPYTER_PASSWORD` from your `.env` file

3. **Create notebooks:**
   - Click "+" to create new notebook
   - Notebooks are auto-saved to this directory

## Pre-installed Libraries

- **Data Science**: NumPy, Pandas, Matplotlib, Seaborn
- **Machine Learning**: Scikit-learn
- **Scientific Computing**: SciPy
- **Web**: Requests
- **Database**: psycopg2 (PostgreSQL), redis

## Example Notebooks

### 1. Connect to PostgreSQL

```python
import psycopg2
import pandas as pd

# Connect to database
conn = psycopg2.connect(
    host="pgvector",
    database="lazy_ai_coder",
    user="postgres",
    password="your_db_password"
)

# Query data
df = pd.read_sql_query("SELECT * FROM users LIMIT 10", conn)
print(df)

conn.close()
```

### 2. Connect to Redis

```python
import redis

# Connect to Redis
r = redis.Redis(
    host='redis',
    port=6379,
    password='your_redis_password',
    decode_responses=True
)

# Test connection
print(r.ping())

# Set/Get values
r.set('test_key', 'Hello from Jupyter!')
print(r.get('test_key'))
```

### 3. Test MCP HTTP API

```python
import requests
import json

# Test MCP server info
response = requests.get('http://lazy-ai-coder:8888/api/v1/mcp/info')
print(json.dumps(response.json(), indent=2))

# List MCP tools
response = requests.get('http://lazy-ai-coder:8888/api/v1/mcp/tools')
tools = response.json()
print(f"Available tools: {tools['count']}")
for tool in tools['tools']:
    print(f"  - {tool['name']}: {tool['description']}")
```

### 4. Call MCP Tools

```python
import requests

# Call LLM chat tool
mcp_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "llm_chat",
        "arguments": {
            "user_prompt": "Explain what is a REST API in one sentence",
            "system_prompt": "You are a helpful assistant."
        }
    }
}

response = requests.post(
    'http://lazy-ai-coder:8888/api/v1/mcp',
    json=mcp_request
)

result = response.json()
print(result['result']['content'][0]['text'])
```

### 5. Data Analysis with Pandas

```python
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

# Create sample data
data = {
    'date': pd.date_range('2024-01-01', periods=100),
    'value': np.random.randn(100).cumsum()
}
df = pd.DataFrame(data)

# Plot
plt.figure(figsize=(12, 6))
plt.plot(df['date'], df['value'])
plt.title('Time Series Data')
plt.xlabel('Date')
plt.ylabel('Value')
plt.grid(True)
plt.show()

# Statistics
print(df.describe())
```

### 6. Machine Learning Example

```python
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, classification_report

# Load data
iris = load_iris()
X_train, X_test, y_train, y_test = train_test_split(
    iris.data, iris.target, test_size=0.2, random_state=42
)

# Train model
clf = RandomForestClassifier(n_estimators=100)
clf.fit(X_train, y_train)

# Evaluate
y_pred = clf.predict(X_test)
print(f"Accuracy: {accuracy_score(y_test, y_pred):.2f}")
print("\nClassification Report:")
print(classification_report(y_test, y_pred, target_names=iris.target_names))
```

### 7. Test GitLab Integration

```python
import requests
import os

# Test GitLab file reading via MCP
mcp_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "get_gitlab_file_content",
        "arguments": {
            "project": "your-org/your-repo",
            "file_path": "README.md",
            "branch": "main"
        }
    }
}

response = requests.post(
    'http://lazy-ai-coder:8888/api/v1/mcp',
    json=mcp_request
)

result = response.json()
if 'result' in result:
    print(result['result']['content'][0]['text'])
```

## Tips

### Install Additional Packages

```python
# In a notebook cell
!pip install package-name

# Or permanently, exec into container:
# docker exec -it jupyterlab pip install package-name
```

### Save Plots

```python
import matplotlib.pyplot as plt

plt.figure()
plt.plot([1, 2, 3], [1, 4, 9])
plt.savefig('/home/jovyan/notebooks/myplot.png')
```

### Environment Variables

```python
import os

# Access env vars
db_password = os.getenv('DB_PASS', 'default')
```

### Widgets for Interactivity

```python
import ipywidgets as widgets
from IPython.display import display

slider = widgets.IntSlider(min=0, max=100, value=50)
display(slider)
```

## Troubleshooting

### Can't connect to other services

**Solution:** Make sure you're using container names:
- PostgreSQL: `host="pgvector"`
- Redis: `host="redis"`
- App: `http://lazy-ai-coder:8888`

### Permission errors

**Solution:** Notebooks run as `jovyan` user. Check file permissions.

### Package not found

**Solution:** Install it:
```python
!pip install package-name
```

## Useful Commands

### Terminal in JupyterLab

Open a terminal from JupyterLab interface to run shell commands:
```bash
# List files
ls -la

# Test network connectivity
ping redis
ping pgvector

# Python version
python --version

# Installed packages
pip list
```

## Best Practices

1. **Save regularly** - Notebooks auto-save, but use Ctrl+S frequently
2. **Clear outputs** - Before committing, clear large outputs
3. **Use markdown** - Document your analysis with markdown cells
4. **Version control** - Commit notebooks to git (without sensitive data)
5. **Keep it organized** - Use subdirectories for different projects

## Resources

- [JupyterLab Documentation](https://jupyterlab.readthedocs.io/)
- [Pandas Documentation](https://pandas.pydata.org/docs/)
- [Scikit-learn Documentation](https://scikit-learn.org/)
- [Matplotlib Gallery](https://matplotlib.org/stable/gallery/)

---

Happy coding! 📊🐍

