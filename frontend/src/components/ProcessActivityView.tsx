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
} from '@mui/material';
import type { ProcessEvent } from '../types';

interface ProcessActivityViewProps {
  events: ProcessEvent[];
}

export const ProcessActivityView = ({ events }: ProcessActivityViewProps) => {
  if (events.length === 0) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          No process activity detected
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer sx={{ maxHeight: 600 }}>
      <Table stickyHeader size="small">
        <TableHead>
          <TableRow>
            <TableCell>Timestamp</TableCell>
            <TableCell>Process Name</TableCell>
            <TableCell>PID</TableCell>
            <TableCell>Command Line</TableCell>
            <TableCell>Child PID</TableCell>
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
                <Chip label={event.process_name} size="small" variant="outlined" />
              </TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {event.pid}
                </Typography>
              </TableCell>
              <TableCell>
                <Typography
                  variant="body2"
                  sx={{
                    fontFamily: 'monospace',
                    fontSize: '0.75rem',
                    maxWidth: 500,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                  title={event.command_line}
                >
                  {event.command_line}
                </Typography>
              </TableCell>
              <TableCell>
                {event.child_pid && (
                  <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                    {event.child_pid}
                  </Typography>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
