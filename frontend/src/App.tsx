import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline, AppBar, Toolbar, Typography, Container, Button } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import BugReportIcon from '@mui/icons-material/BugReport';
import { FileUpload } from './components/FileUpload';
import { Dashboard } from './pages/Dashboard';
import { AnalysisReport } from './pages/AnalysisReport';
import { WhitelistPage } from './pages/Whitelist';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <AppBar position="static">
          <Toolbar>
            <BugReportIcon sx={{ mr: 2 }} />
            <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
              CupaX - Malware Analysis Platform
            </Typography>
            <Button color="inherit" component={RouterLink} to="/upload">
              Upload
            </Button>
            <Button color="inherit" component={RouterLink} to="/dashboard">
              Dashboard
            </Button>
            <Button color="inherit" component={RouterLink} to="/whitelist">
              Whitelist
            </Button>
          </Toolbar>
        </AppBar>

        <Container>
          <Routes>
            <Route path="/" element={<Navigate to="/upload" replace />} />
            <Route path="/upload" element={<FileUpload />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/analysis/:id" element={<AnalysisReport />} />
            <Route path="/whitelist" element={<WhitelistPage />} />
          </Routes>
        </Container>
      </Router>
    </ThemeProvider>
  );
}

export default App;
