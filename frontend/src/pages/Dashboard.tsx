import { useState, useEffect } from 'react';
import {
  Box,
  Container,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Chip,
  IconButton,
  CircularProgress,
  Alert,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import VisibilityIcon from '@mui/icons-material/Visibility';
import { useNavigate } from 'react-router-dom';
import { getAnalyses } from '../services/api';
import type { Analysis, AnalysisStatus } from '../types';

const getStatusColor = (status: AnalysisStatus) => {
  switch (status) {
    case 'running':
      return 'info';
    case 'completed':
      return 'success';
    case 'error':
      return 'error';
    default:
      return 'default';
  }
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString();
};

export const Dashboard = () => {
  const [analyses, setAnalyses] = useState<Analysis[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const fetchAnalyses = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getAnalyses();
      setAnalyses(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch analyses');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAnalyses();

    // Poll for updates every 5 seconds
    const interval = setInterval(fetchAnalyses, 5000);

    return () => clearInterval(interval);
  }, []);

  const handleViewReport = (id: string) => {
    navigate(`/analysis/${id}`);
  };

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Analysis Dashboard
        </Typography>
        <IconButton onClick={fetchAnalyses} disabled={loading}>
          <RefreshIcon />
        </IconButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {loading && analyses.length === 0 ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Filename</TableCell>
                <TableCell>SHA256 Hash</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Submitted</TableCell>
                <TableCell>Completed</TableCell>
                <TableCell align="center">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {analyses.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} align="center">
                    <Typography variant="body2" color="text.secondary">
                      No analyses found. Upload a file to get started.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                analyses.map((analysis) => (
                  <TableRow key={analysis.id} hover>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                        {analysis.filename}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                        {analysis.file_hash_sha256.substring(0, 16)}...
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={analysis.status.toUpperCase()}
                        color={getStatusColor(analysis.status)}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {formatDate(analysis.submitted_at)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {analysis.completed_at ? formatDate(analysis.completed_at) : '-'}
                      </Typography>
                    </TableCell>
                    <TableCell align="center">
                      <IconButton
                        size="small"
                        onClick={() => handleViewReport(analysis.id)}
                        disabled={analysis.status !== 'completed'}
                      >
                        <VisibilityIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Container>
  );
};
