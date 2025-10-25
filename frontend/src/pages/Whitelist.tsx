import { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
  Alert,
  Paper,
  Tabs,
  Tab,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
} from '@mui/icons-material';
import {
  getWhitelists,
  createWhitelist,
  updateWhitelist,
  deleteWhitelist,
} from '../services/api';
import type { Whitelist, WhitelistCreate, WhitelistType } from '../types/whitelist';

export const WhitelistPage = () => {
  const [whitelists, setWhitelists] = useState<Whitelist[]>([]);
  const [filteredWhitelists, setFilteredWhitelists] = useState<Whitelist[]>([]);
  const [selectedTab, setSelectedTab] = useState<WhitelistType | 'all'>('all');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Dialog states
  const [openDialog, setOpenDialog] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState<WhitelistCreate>({
    type: 'process',
    value: '',
    description: '',
    is_regex: false,
    enabled: true,
  });

  // Load whitelists
  const loadWhitelists = async () => {
    setError(null);
    try {
      const data = await getWhitelists();
      setWhitelists(data);
      filterWhitelists(data, selectedTab);
    } catch (err: any) {
      setError('Failed to load whitelists');
    }
  };

  useEffect(() => {
    loadWhitelists();
    const interval = setInterval(loadWhitelists, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  // Filter whitelists by type
  const filterWhitelists = (data: Whitelist[], type: WhitelistType | 'all') => {
    if (type === 'all') {
      setFilteredWhitelists(data);
    } else {
      setFilteredWhitelists(data.filter(w => w.type === type));
    }
  };

  const handleTabChange = (_: any, newValue: WhitelistType | 'all') => {
    setSelectedTab(newValue);
    filterWhitelists(whitelists, newValue);
  };

  // Open dialog for create/edit
  const handleOpenDialog = (whitelist?: Whitelist) => {
    if (whitelist) {
      setEditingId(whitelist.id);
      setFormData({
        type: whitelist.type,
        value: whitelist.value,
        description: whitelist.description,
        is_regex: whitelist.is_regex,
        enabled: whitelist.enabled,
      });
    } else {
      setEditingId(null);
      setFormData({
        type: selectedTab !== 'all' ? selectedTab : 'process',
        value: '',
        description: '',
        is_regex: false,
        enabled: true,
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditingId(null);
    setFormData({
      type: 'process',
      value: '',
      description: '',
      is_regex: false,
      enabled: true,
    });
  };

  // Handle form submission
  const handleSubmit = async () => {
    setError(null);
    setSuccess(null);
    try {
      if (editingId) {
        await updateWhitelist(editingId, formData);
        setSuccess('Whitelist updated successfully');
      } else {
        await createWhitelist(formData);
        setSuccess('Whitelist created successfully');
      }
      handleCloseDialog();
      loadWhitelists();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save whitelist');
    }
  };

  // Handle delete
  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this whitelist entry?')) return;

    setError(null);
    setSuccess(null);
    try {
      await deleteWhitelist(id);
      setSuccess('Whitelist deleted successfully');
      loadWhitelists();
    } catch (err: any) {
      setError('Failed to delete whitelist');
    }
  };

  // Handle toggle enabled
  const handleToggleEnabled = async (whitelist: Whitelist) => {
    try {
      await updateWhitelist(whitelist.id, { enabled: !whitelist.enabled });
      loadWhitelists();
    } catch (err: any) {
      setError('Failed to update whitelist');
    }
  };

  const getTypeColor = (type: WhitelistType) => {
    switch (type) {
      case 'process':
        return 'primary';
      case 'domain':
        return 'secondary';
      case 'ip':
        return 'success';
      case 'registry':
        return 'warning';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Whitelist Management</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
        >
          Add Whitelist
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="body2" color="text.secondary">
            Whitelists reduce noise by filtering out benign processes, domains, IPs, and registry keys from analysis reports.
            Enable/disable entries to control filtering behavior.
          </Typography>
        </CardContent>
      </Card>

      <Tabs value={selectedTab} onChange={handleTabChange} sx={{ mb: 2 }}>
        <Tab label={`All (${whitelists.length})`} value="all" />
        <Tab label={`Processes (${whitelists.filter(w => w.type === 'process').length})`} value="process" />
        <Tab label={`Domains (${whitelists.filter(w => w.type === 'domain').length})`} value="domain" />
        <Tab label={`IPs (${whitelists.filter(w => w.type === 'ip').length})`} value="ip" />
        <Tab label={`Registry (${whitelists.filter(w => w.type === 'registry').length})`} value="registry" />
      </Tabs>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Type</TableCell>
              <TableCell>Value</TableCell>
              <TableCell>Description</TableCell>
              <TableCell>Regex</TableCell>
              <TableCell>Enabled</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredWhitelists.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  <Typography variant="body2" color="text.secondary">
                    No whitelists found. Click "Add Whitelist" to create one.
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              filteredWhitelists.map((whitelist) => (
                <TableRow key={whitelist.id}>
                  <TableCell>
                    <Chip
                      label={whitelist.type.toUpperCase()}
                      color={getTypeColor(whitelist.type)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {whitelist.value}
                    </Typography>
                  </TableCell>
                  <TableCell>{whitelist.description || '-'}</TableCell>
                  <TableCell>
                    {whitelist.is_regex ? (
                      <Chip label="Regex" size="small" variant="outlined" />
                    ) : (
                      '-'
                    )}
                  </TableCell>
                  <TableCell>
                    <Switch
                      checked={whitelist.enabled}
                      onChange={() => handleToggleEnabled(whitelist)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell align="right">
                    <IconButton
                      size="small"
                      onClick={() => handleOpenDialog(whitelist)}
                      color="primary"
                    >
                      <EditIcon />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDelete(whitelist.id)}
                      color="error"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Create/Edit Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{editingId ? 'Edit Whitelist' : 'Add Whitelist'}</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>Type</InputLabel>
              <Select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as WhitelistType })}
                label="Type"
              >
                <MenuItem value="process">Process</MenuItem>
                <MenuItem value="domain">Domain</MenuItem>
                <MenuItem value="ip">IP Address</MenuItem>
                <MenuItem value="registry">Registry Key</MenuItem>
              </Select>
            </FormControl>

            <TextField
              fullWidth
              label="Value"
              value={formData.value}
              onChange={(e) => setFormData({ ...formData, value: e.target.value })}
              placeholder={
                formData.type === 'process'
                  ? 'e.g., svchost.exe'
                  : formData.type === 'domain'
                  ? 'e.g., microsoft.com'
                  : formData.type === 'ip'
                  ? 'e.g., 192.168.1.1 or ^192\\.168\\..*'
                  : 'e.g., HKLM\\SOFTWARE\\Microsoft'
              }
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Optional description"
              sx={{ mb: 2 }}
            />

            <Box sx={{ display: 'flex', gap: 2 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.is_regex}
                    onChange={(e) => setFormData({ ...formData, is_regex: e.target.checked })}
                  />
                }
                label="Use Regex"
              />

              <FormControlLabel
                control={
                  <Switch
                    checked={formData.enabled}
                    onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  />
                }
                label="Enabled"
              />
            </Box>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">
            {editingId ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
