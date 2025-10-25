import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Container,
  Paper,
  Typography,
  Card,
  CardContent,
  Tabs,
  Tab,
  CircularProgress,
  Alert,
  IconButton,
  Grid,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import BugReportIcon from '@mui/icons-material/BugReport';
import FolderIcon from '@mui/icons-material/Folder';
import SettingsIcon from '@mui/icons-material/Settings';
import NetworkCheckIcon from '@mui/icons-material/NetworkCheck';
import { getAnalysisById } from '../services/api';
import type { Analysis } from '../types';
import { ProcessActivityView } from '../components/ProcessActivityView';
import { FileSystemView } from '../components/FileSystemView';
import { RegistryView } from '../components/RegistryView';
import { NetworkView } from '../components/NetworkView';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

const TabPanel = (props: TabPanelProps) => {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tabpanel-${index}`}
      aria-labelledby={`tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
};

export const AnalysisReport = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    const fetchAnalysis = async () => {
      if (!id) return;

      try {
        setLoading(true);
        const data = await getAnalysisById(id);
        setAnalysis(data);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch analysis');
      } finally {
        setLoading(false);
      }
    };

    fetchAnalysis();
  }, [id]);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !analysis) {
    return (
      <Container maxWidth="xl" sx={{ mt: 4 }}>
        <Alert severity="error">{error || 'Analysis not found'}</Alert>
      </Container>
    );
  }

  const report = analysis.report_json;

  if (!report) {
    return (
      <Container maxWidth="xl" sx={{ mt: 4 }}>
        <Alert severity="warning">Analysis report not yet available</Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <IconButton onClick={() => navigate('/dashboard')} sx={{ mr: 2 }}>
          <ArrowBackIcon />
        </IconButton>
        <Box>
          <Typography variant="h4" component="h1">
            Analysis Report
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            {analysis.filename}
          </Typography>
        </Box>
      </Box>

      {/* Summary Statistics */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" color="primary">
                    {report.summary.processes_created || 0}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Processes Created
                  </Typography>
                </Box>
                <BugReportIcon sx={{ fontSize: 40, color: 'primary.main', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" color="success.main">
                    {report.summary.files_created || 0}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Files Created
                  </Typography>
                </Box>
                <FolderIcon sx={{ fontSize: 40, color: 'success.main', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" color="warning.main">
                    {report.summary.registry_modified || 0}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Registry Modifications
                  </Typography>
                </Box>
                <SettingsIcon sx={{ fontSize: 40, color: 'warning.main', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" color="error.main">
                    {report.summary.network_connections || 0}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Network Connections
                  </Typography>
                </Box>
                <NetworkCheckIcon sx={{ fontSize: 40, color: 'error.main', opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Timing Information */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Execution Details
        </Typography>
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Typography variant="body2" color="text.secondary">
              Execution Time
            </Typography>
            <Typography variant="body1">
              {report.summary.execution_time?.toFixed(2) || 0}s
            </Typography>
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Typography variant="body2" color="text.secondary">
              Processing Time
            </Typography>
            <Typography variant="body1">
              {report.summary.processing_time?.toFixed(2) || 0}s
            </Typography>
          </Grid>
          <Grid size={{ xs: 12, sm: 4 }}>
            <Typography variant="body2" color="text.secondary">
              Analysis Time
            </Typography>
            <Typography variant="body1">
              {report.summary.analysis_time?.toFixed(2) || 0}s
            </Typography>
          </Grid>
        </Grid>
      </Paper>

      {/* Detailed Tabs */}
      <Paper sx={{ width: '100%' }}>
        <Tabs value={tabValue} onChange={handleTabChange} aria-label="analysis tabs">
          <Tab label={`Processes (${report.process_activity.length})`} />
          <Tab label={`File System (${report.file_system.length})`} />
          <Tab label={`Registry (${report.registry.length})`} />
          <Tab label={`Network (${report.network.length})`} />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          <ProcessActivityView events={report.process_activity} />
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <FileSystemView events={report.file_system} />
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          <RegistryView events={report.registry} />
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          <NetworkView events={report.network} uniqueHosts={report.unique_hosts} />
        </TabPanel>
      </Paper>
    </Container>
  );
};
