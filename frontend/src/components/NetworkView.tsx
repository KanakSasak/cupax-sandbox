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
  Paper,
  List,
  ListItem,
  ListItemText,
  Grid,
} from '@mui/material';
import type { NetworkEvent } from '../types';

interface NetworkViewProps {
  events: NetworkEvent[];
  uniqueHosts: string[];
}

const getProtocolColor = (protocol: string) => {
  switch (protocol.toUpperCase()) {
    case 'TCP':
      return 'primary';
    case 'UDP':
      return 'secondary';
    default:
      return 'default';
  }
};

const getDirectionColor = (direction: string) => {
  switch (direction.toLowerCase()) {
    case 'send':
      return 'error';
    case 'receive':
      return 'success';
    default:
      return 'default';
  }
};

export const NetworkView = ({ events, uniqueHosts }: NetworkViewProps) => {
  if (events.length === 0 && uniqueHosts.length === 0) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          No network activity detected
        </Typography>
      </Box>
    );
  }

  return (
    <Grid container spacing={3}>
      <Grid size={{ xs: 12, md: 8 }}>
        <TableContainer sx={{ maxHeight: 600 }}>
          <Table stickyHeader size="small">
            <TableHead>
              <TableRow>
                <TableCell>Timestamp</TableCell>
                <TableCell>Protocol</TableCell>
                <TableCell>Direction</TableCell>
                <TableCell>Process</TableCell>
                <TableCell>PID</TableCell>
                <TableCell>Remote Address</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {events.map((event, index) => (
                <TableRow key={index} hover>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                      {event.timestamp}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={event.protocol}
                      size="small"
                      color={getProtocolColor(event.protocol)}
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={event.direction}
                      size="small"
                      color={getDirectionColor(event.direction)}
                      variant="outlined"
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
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                      {event.remote_addr}
                    </Typography>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Grid>

      <Grid size={{ xs: 12, md: 4 }}>
        <Paper sx={{ p: 2, maxHeight: 600, overflow: 'auto' }}>
          <Typography variant="h6" gutterBottom>
            Unique Hosts ({uniqueHosts.length})
          </Typography>
          <List dense>
            {uniqueHosts.map((host, index) => (
              <ListItem key={index}>
                <ListItemText
                  primary={
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {host}
                    </Typography>
                  }
                />
              </ListItem>
            ))}
          </List>
        </Paper>
      </Grid>
    </Grid>
  );
};
