/**
 * Grouped Releases Module - Handle toggle and copy functionality
 */

// Use event delegation to handle all grouped release card interactions
document.addEventListener('click', async (e: Event) => {
  const target = e.target as HTMLElement;
  
  // Handle copy button clicks
  const copyButton = target.closest('.copy-button') as HTMLButtonElement | null;
  if (copyButton) {
    const command = copyButton.getAttribute('data-command');
    if (!command) return;
    
    try {
      await navigator.clipboard.writeText(command);
      const icon = copyButton.querySelector('i');
      if (icon) {
        icon.className = 'fas fa-check';
        copyButton.classList.add('copied');
        setTimeout(() => {
          icon.className = 'fas fa-copy';
          copyButton.classList.remove('copied');
        }, 2000);
      }
    } catch (err) {
      console.error('Failed to copy:', err);
    }
    return;
  }
  
  // Handle toggle button clicks
  const toggleButton = target.closest('.older-releases-toggle') as HTMLButtonElement | null;
  if (toggleButton) {
    const groupId = toggleButton.getAttribute('data-group-id');
    if (!groupId) return;
    
    const list = document.getElementById(groupId);
    const isExpanded = toggleButton.getAttribute('aria-expanded') === 'true';
    
    toggleButton.setAttribute('aria-expanded', (!isExpanded).toString());
    if (list) {
      list.hidden = isExpanded;
    }
    
    // Update toggle text
    const toggleText = toggleButton.querySelector('.toggle-text');
    if (toggleText) {
      const releaseCount = list?.querySelectorAll('.older-release-item').length || 0;
      toggleText.textContent = isExpanded 
        ? `Show ${releaseCount} older release${releaseCount !== 1 ? 's' : ''}`
        : `Hide older releases`;
    }
  }
});

console.log('[GroupedReleases] Event delegation initialized');
