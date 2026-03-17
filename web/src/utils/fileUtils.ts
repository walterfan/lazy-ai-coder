export function getFileExtension(url: string): string {
  try {
    const urlObj = new URL(url, window.location.href);
    const pathname = urlObj.pathname;
    const parts = pathname.split('.');
    return parts.length > 1 ? `.${parts.pop()}` : '';
  } catch {
    return '';
  }
}

export function downloadFile(
  content: string,
  filename: string,
  contentType: string = 'text/markdown'
): void {
  const blob = new Blob([content], { type: contentType });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export async function copyToClipboard(
  content: string,
  successMessage: string = 'Content copied to clipboard successfully!'
): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(content);
    const { showToast } = await import('./toast');
    showToast(successMessage, 'success');
    return true;
  } catch (err) {
    const { showToast } = await import('./toast');
    const errorMessage = err instanceof Error ? err.message : 'Unknown error';
    showToast(`Failed to copy to clipboard: ${errorMessage}`, 'danger');
    return false;
  }
}
