// Expose as globals to be used by inline scripts in templates.
window.renderStatusBadge = function renderStatusBadge(status) {
    const s = (status || '').toString().toLowerCase();
    let cls = 'badge-info';
    let icon = '';
    const label = s || 'unknown';

    switch (s) {
        case 'completed':
            cls = 'badge-completed';
            icon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="currentColor" d="M256 48C141.31 48 48 141.31 48 256s93.31 208 208 208s208-93.31 208-208S370.69 48 256 48m108.25 138.29l-134.4 160a16 16 0 0 1-12 5.71h-.27a16 16 0 0 1-11.89-5.3l-57.6-64a16 16 0 1 1 23.78-21.4l45.29 50.32l122.59-145.91a16 16 0 0 1 24.5 20.58"/></svg>';
            break;
        case 'failed':
        case 'download_failed':
        case 'transcription_failed':
        case 'summary_failed':
        case 'metadata_failed':
            cls = 'badge-error';
            icon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path d="M256 48C141.1 48 48 141.1 48 256s93.1 208 208 208 208-93.1 208-208S370.9 48 256 48zm52.7 283.3L256 278.6l-52.7 52.7c-6.2 6.2-16.4 6.2-22.6 0-3.1-3.1-4.7-7.2-4.7-11.3 0-4.1 1.6-8.2 4.7-11.3l52.7-52.7-52.7-52.7c-3.1-3.1-4.7-7.2-4.7-11.3 0-4.1 1.6-8.2 4.7-11.3 6.2-6.2 16.4-6.2 22.6 0l52.7 52.7 52.7-52.7c6.2-6.2 16.4-6.2 22.6 0 6.2 6.2 6.2 16.4 0 22.6L278.6 256l52.7 52.7c6.2 6.2 6.2 16.4 0 22.6-6.2 6.3-16.4 6.3-22.6 0z" fill="currentColor"/></svg>';
            break;
        case 'downloading':
        case 'transcribing':
        case 'summarizing':
        case 'fetching_metadata':
        case 'processing':
            cls = 'badge-progress';
            icon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path d="M128 48v122.8h.2l-.2.2 85.3 85-85.3 85.2.2.2h-.2V464h256V341.4h-.2l.2-.2-85.3-85.2 85.3-85-.2-.2h.2V48H128zm213.3 303.9v71.5H170.7v-71.5l85.3-85.2 85.3 85.2zM256 245.4l-85.3-85.2V87.6h170.7v72.5L256 245.4z" fill="currentColor"/></svg>';
            break;
        case 'pending':
            cls = 'badge-pending';
            icon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path d="M256 48C141.1 48 48 141.1 48 256s93.1 208 208 208 208-93.1 208-208S370.9 48 256 48zm0 398.7c-105.1 0-190.7-85.5-190.7-190.7 0-105.1 85.5-190.7 190.7-190.7 105.1 0 190.7 85.5 190.7 190.7 0 105.1-85.6 190.7-190.7 190.7z" fill="currentColor"/><path d="M256 256h-96v17.3h113.3V128H256z" fill="currentColor"/></svg>';
            break;
        default:
            cls = 'badge-info';
            icon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 512 512"><path fill="currentColor" d="M256 56C145.72 56 56 145.72 56 256s89.72 200 200 200s200-89.72 200-200S366.28 56 256 56m0 82a26 26 0 1 1-26 26a26 26 0 0 1 26-26m64 226H200v-32h44v-88h-32v-32h64v120h44Z"/></svg>';
    }

    const pretty = label.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
    return `<span class="badge ${cls}">${icon}<span>${pretty}</span></span>`;
};

window.formatUploadDate = function formatUploadDate(dateStr) {
    if (dateStr && dateStr.length === 8) {
        // Assuming YYYYMMDD format
        return `${dateStr.substring(0, 4)}-${dateStr.substring(4, 6)}-${dateStr.substring(6, 8)}`;
    }
    return dateStr; // Return original if not in expected format
}
