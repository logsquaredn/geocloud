import React from "react";
import { Alert, AlertColor, AppBar, Avatar, Box, Container, CssBaseline, Grid, TextField, Typography, ThemeProvider, Toolbar, createTheme, Button, Link } from "@mui/material";
import { LockOutlined } from "@mui/icons-material";

function createAPIKey(email: string) {
  return fetch("/api/v1/api-key", {
      method: "POST",
      body: JSON.stringify({ email }),
  });
}

const theme = createTheme();

function App() {
  const [alert, setAlert] = React.useState<{ msg?: string, severity?: AlertColor }>({});
  const [apiKey, setAPIKey] = React.useState<string | undefined>();

  const handleCopy = () => {
    if (apiKey) {
      navigator.clipboard.writeText(apiKey);
    }

    setAlert({ msg: "Copied", severity: "success" });
    setAPIKey(undefined);
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const email = new FormData(event.currentTarget).get("email")?.toString();
    if (email) {
      createAPIKey(email).then(async res => {
        switch (res.status) {
          case 200:
            res.json().then(body => {
              if (body.api_key) {
                setAPIKey(body.api_key);
                setAlert({ msg: `Click to copy API key`, severity: "info" });
              } else {
                setAlert({ msg: "Something went wrong. Try again", severity: "error" });
              }
            })
            break;
          case 201:
            setAlert({ msg: `API key emailed to ${email}` });
            break;
          default:
            res.json().then(body => {
              setAlert({ msg: body.error, severity: "error" });
            }).catch(() => {
              setAlert({ msg: res.statusText, severity: "error" });
            })
        }
      });
    }
  };

  const handleInvalid = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setAlert({ msg: "Invalid email address", severity: "error" });
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar
        position="static"
        color="default"
        elevation={0}
        sx={{
          position: "relative",
          borderBottom: (t) => `1px solid ${t.palette.divider}`,
        }}
      >
        <Toolbar sx={{ flexWrap: "wrap" }}>
          <Typography variant="h6" color="inherit" sx={{ flexGrow: 1 }}>
            Rototiller
          </Typography>
          <nav>
            <Link
              variant="button"
              color="text.primary"
              href="/swagger/v1"
              sx={{ my: 1, mx: 1.5 }}
            >
              Docs
            </Link>
          </nav>
        </Toolbar>
      </AppBar>
      <Container component="main" maxWidth="xs">
        <Box
          sx={{
            marginTop: 8,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <Avatar sx={{ m: 1, bgcolor: "secondary.main" }}>
            <LockOutlined />
          </Avatar>
          <Typography component="h1" variant="h5">
            Get API key
          </Typography>
          <Box component="form" onSubmit={handleSubmit} onInvalid={handleInvalid} sx={{ mt: 3 }}>
            <Grid container spacing={2}>
              <Grid item>
                <TextField
                  required
                  fullWidth
                  id="email"
                  label="Email Address"
                  name="email"
                  autoComplete="email"
                />
              </Grid>
            </Grid>
            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
            >
              Submit
            </Button>
          </Box>
          {alert.msg &&
            <Alert
              severity={alert.severity}
              onClick={apiKey ? handleCopy : undefined}
              style={{
                cursor: apiKey? "pointer" : undefined,
              }}
            >
              {alert.msg}
            </Alert>
          }
        </Box>
      </Container>
    </ThemeProvider>
  );
}

export default App;
