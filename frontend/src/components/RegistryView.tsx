import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Box,
  Chip,
  TextField,
  InputAdornment,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import type { RegistryEvent } from '../types';

interface RegistryViewProps {
  events: RegistryEvent[];
}

const getOperationColor = (operation: string) => {
  switch (operation.toLowerCase()) {
    case 'regcreatekey':
      return 'success';
    case 'regsetvalue':
      return 'info';
    case 'regdeletevalue':
    case 'regdeletekey':
      return 'error';
    default:
      return 'default';
  }
};

export const RegistryView = ({ events }: RegistryViewProps) => {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredEvents = events.filter(
    (event) =>
      event.path.toLowerCase().includes(searchTerm.toLowerCase()) ||
      event.operation.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (events.length === 0) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          No registry activity detected
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ p: 2, pb: 1 }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Search registry paths or operations..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      <TableContainer sx={{ maxHeight: 600 }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell>Operation</TableCell>
              <TableCell>Process</TableCell>
              <TableCell>PID</TableCell>
              <TableCell>Registry Path</TableCell>
              <TableCell>Data</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredEvents.map((event, index) => (
              <TableRow key={index} hover>
                <TableCell>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                    {event.timestamp}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={event.operation}
                    size="small"
                    color={getOperationColor(event.operation)}
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                    {event.process_name}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                    {event.pid}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography
                    variant="body2"
                    sx={{
                      fontFamily: 'monospace',
                      fontSize: '0.75rem',
                      maxWidth: 400,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                    title={event.path}
                  >
                    {event.path}
                  </Typography>
                </TableCell>
                <TableCell>
                  {event.data && (
                    <Typography
                      variant="body2"
                      sx={{
                        fontFamily: 'monospace',
                        fontSize: '0.75rem',
                        maxWidth: 300,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                      title={event.data}
                    >
                      {event.data}
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Box sx={{ p: 2, textAlign: 'center' }}>
        <Typography variant="caption" color="text.secondary">
          Showing {filteredEvents.length} of {events.length} events
        </Typography>
      </Box>
    </Box>
  );
};
