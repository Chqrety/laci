/**
 * LACI DOSCOM - Google Apps Script Webhook
 * 
 * Panduan Deployment:
 * 1. Buka https://script.google.com/
 * 2. Buat Project Baru: "Laci Generator Surat"
 * 3. Hapus semua kode default, paste seluruh kode ini
 * 4. Klik "Deploy" -> "New deployment"
 * 5. Pilih tipe: "Web app"
 * 6. Set:
 *    - Description: "Laci Surat Webhook"
 *    - Execute as: "Me" (Email pemilik Google Drive)
 *    - Who has access: "Anyone"
 * 7. Klik "Deploy", beri izin (Authorize access), lalu salin "Web app URL"
 * 8. Masukkan Web app URL tersebut ke menu "Setup Template" di aplikasi Laci
 */

function doPost(e) {
  try {
    if (!e || !e.postData || !e.postData.contents) {
      return jsonResponse({ status: 'error', message: 'Payload tidak ditemukan' }, 400);
    }

    var payload = JSON.parse(e.postData.contents);
    var action = payload.action;

    if (action === 'sync_templates') {
      return handleSyncTemplates(payload);
    } else if (action === 'generate_surat') {
      return handleGenerateSurat(payload);
    } else {
      return jsonResponse({ status: 'error', message: 'Action tidak dikenal: ' + action }, 400);
    }
  } catch (err) {
    return jsonResponse({ status: 'error', message: err.toString() }, 500);
  }
}

function handleSyncTemplates(payload) {
  var folderId = extractId(payload.folder_id || payload.folder_url);
  if (!folderId) {
    return jsonResponse({ status: 'error', message: 'Folder ID atau URL Google Drive tidak valid' }, 400);
  }

  var folder = DriveApp.getFolderById(folderId);
  var files = folder.getFilesByType(MimeType.GOOGLE_DOCS);
  var templates = [];

  while (files.hasNext()) {
    var file = files.next();
    var doc = DocumentApp.openById(file.getId());
    var bodyText = doc.getBody().getText();

    // Deteksi seluruh placeholder dengan format {{variabel}}
    var regex = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g;
    var placeholders = [];
    var match;
    var seen = {};

    while ((match = regex.exec(bodyText)) !== null) {
      var key = match[1].trim();
      if (!seen[key]) {
        seen[key] = true;
        placeholders.push(key);
      }
    }

    templates.push({
      gdoc_id: file.getId(),
      nama: file.getName(),
      placeholders: placeholders
    });
  }

  return jsonResponse({
    status: 'success',
    templates: templates,
    count: templates.length
  });
}

function handleGenerateSurat(payload) {
  var templateId = extractId(payload.template_id || payload.template_url);
  var targetFolderId = extractId(payload.folder_id || payload.folder_url);
  var fileName = payload.file_name || 'Surat Keluar DOSCOM';
  var replacements = payload.replacements || {};

  if (!templateId) {
    return jsonResponse({ status: 'error', message: 'Template ID wajib diisi' }, 400);
  }

  var templateFile = DriveApp.getFileById(templateId);
  var targetFolder = targetFolderId ? DriveApp.getFolderById(targetFolderId) : DriveApp.getRootFolder();

  // 1. Duplikasi file template ke folder tujuan
  var newFile = templateFile.makeCopy(fileName, targetFolder);
  var doc = DocumentApp.openById(newFile.getId());
  var body = doc.getBody();

  // 2. Ganti semua placeholder {{...}} dengan teks aktual
  for (var key in replacements) {
    if (Object.prototype.hasOwnProperty.call(replacements, key)) {
      var rawVal = replacements[key];
      var strVal = (rawVal === null || rawVal === undefined) ? '' : String(rawVal);
      var pattern = '\\{\\{\\s*' + escapeRegex(key) + '\\s*\\}\\}';
      body.replaceText(pattern, strVal);
    }
  }

  doc.saveAndClose();

  var docId = newFile.getId();
  return jsonResponse({
    status: 'success',
    doc_id: docId,
    gdocs_url: 'https://docs.google.com/document/d/' + docId + '/edit',
    pdf_url: 'https://docs.google.com/document/d/' + docId + '/export?format=pdf'
  });
}

function extractId(urlOrId) {
  if (!urlOrId) return '';
  var str = String(urlOrId).trim();
  var match = str.match(/[-\w]{25,}/);
  return match ? match[0] : str;
}

function escapeRegex(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function jsonResponse(data, statusCode) {
  return ContentService.createTextOutput(JSON.stringify(data))
    .setMimeType(ContentService.MimeType.JSON);
}
