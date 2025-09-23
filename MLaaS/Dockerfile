FROM python:3.10-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 5555

#CMD ["python", "api.py"]
CMD ["uvicorn", "api:app", "--host", "0.0.0.0", "--port", "5555"]
