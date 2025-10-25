import { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import {
  Box,
  Paper,
  Typography,
  Button,
  LinearProgress,
  Alert,
  Stack,
  TextField,
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { uploadFile } from '../services/api';
import { useNavigate } from 'react-router-dom';

export const FileUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [zipPassword, setZipPassword] = useState<string>('infected');
  const navigate = useNavigate();

  const onDrop = useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) return;

    const file = acceptedFiles[0];
    setSelectedFile(file);
    setError(null);
    setSuccess(null);
  }, []);

  const handleUpload = async () => {
    if (!selectedFile) return;

    setError(null);
    setSuccess(null);
    setUploading(true);

    try {
      const isZip = selectedFile.name.toLowerCase().endsWith('.zip');
      const response = await uploadFile(selectedFile, isZip ? zipPassword : undefined);
      setSuccess(response.message);

      // Redirect to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to upload file');
    } finally {
      setUploading(false);
    }
  };

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: false,
    maxSize: 100 * 1024 * 1024, // 100MB
  });

  return (
    <Box sx={{ maxWidth: 600, margin: 'auto', mt: 8 }}>
      <Typography variant="h4" gutterBottom align="center" sx={{ mb: 4 }}>
        Upload Malware Sample
      </Typography>

      <Paper
        {...getRootProps()}
        sx={{
          p: 6,
          textAlign: 'center',
          cursor: 'pointer',
          border: '2px dashed',
          borderColor: isDragActive ? 'primary.main' : 'grey.300',
          bgcolor: isDragActive ? 'action.hover' : 'background.paper',
          transition: 'all 0.3s',
          '&:hover': {
            borderColor: 'primary.main',
            bgcolor: 'action.hover',
          },
        }}
      >
        <input {...getInputProps()} />
        <Stack spacing={2} alignItems="center">
          <CloudUploadIcon sx={{ fontSize: 64, color: 'primary.main' }} />
          <Typography variant="h6">
            {isDragActive ? 'Drop the file here' : 'Drag and drop a file here'}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            or
          </Typography>
          <Button variant="contained" component="span">
            Browse Files
          </Button>
          <Typography variant="caption" color="text.secondary">
            Maximum file size: 100MB
          </Typography>
        </Stack>
      </Paper>

      {selectedFile && (
        <Box sx={{ mt: 3 }}>
          <Alert severity="info" sx={{ mb: 2 }}>
            Selected file: {selectedFile.name}
          </Alert>

          {selectedFile.name.toLowerCase().endsWith('.zip') && (
            <TextField
              fullWidth
              label="ZIP Password"
              value={zipPassword}
              onChange={(e) => setZipPassword(e.target.value)}
              placeholder="infected"
              helperText="Default password is 'infected'"
              sx={{ mb: 2 }}
            />
          )}

          <Button
            variant="contained"
            fullWidth
            onClick={handleUpload}
            disabled={uploading}
          >
            Upload and Analyze
          </Button>
        </Box>
      )}

      {uploading && (
        <Box sx={{ mt: 3 }}>
          <LinearProgress />
          <Typography variant="body2" align="center" sx={{ mt: 1 }}>
            Uploading and queuing analysis...
          </Typography>
        </Box>
      )}

      {error && (
        <Alert severity="error" sx={{ mt: 3 }}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mt: 3 }}>
          {success}
        </Alert>
      )}
    </Box>
  );
};
