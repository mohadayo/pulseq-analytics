import app from "./app";

const PORT = process.env.PORT || 8003;

app.listen(PORT, () => {
  console.log(`Dashboard BFF listening on port ${PORT}`);
});
