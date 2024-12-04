from fastapi import FastAPI
from fastapi.responses import JSONResponse
import random
import string

app = FastAPI()
large_data = []


@app.get("/health")
async def health_check():
    global large_data
    large_string = ''.join(random.choices(string.ascii_letters + string.digits, k=10**5))
    large_data.append(large_string)
    return JSONResponse(content={"mem": 256, "cpu": 20}, status_code=200)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
