import { Toast } from 'bootstrap';
import type { ToastType } from '@/types';

export function showToast(message: string, type: ToastType = 'success'): void {
  const toastContainer = document.getElementById('toast-container') || createToastContainer();

  const toast = document.createElement('div');
  toast.className = `toast align-items-center text-white bg-${type} border-0`;
  toast.setAttribute('role', 'alert');
  toast.setAttribute('aria-live', 'assertive');
  toast.setAttribute('aria-atomic', 'true');

  const iconClass = type === 'success'
    ? 'fa-check-circle'
    : type === 'danger'
    ? 'fa-exclamation-triangle'
    : type === 'warning'
    ? 'fa-exclamation-circle'
    : 'fa-info-circle';

  toast.innerHTML = `
    <div class="d-flex">
      <div class="toast-body" style="font-size: 1.1rem;">
        <i class="fas ${iconClass}"></i>
        ${message}
      </div>
      <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
    </div>
  `;

  toastContainer.appendChild(toast);
  const bsToast = new Toast(toast);
  bsToast.show();

  toast.addEventListener('hidden.bs.toast', () => {
    toast.remove();
  });
}

function createToastContainer(): HTMLElement {
  const container = document.createElement('div');
  container.id = 'toast-container';
  container.className = 'toast-container position-fixed top-0 start-50 translate-middle-x p-3';
  container.style.zIndex = '1055';
  document.body.appendChild(container);
  return container;
}
