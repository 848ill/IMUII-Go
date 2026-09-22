import os
import re
import docx
from docx import Document
from docx.shared import Pt, RGBColor, Cm
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT, WD_ALIGN_VERTICAL
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import nsdecls, qn

def create_base_doc():
    doc = Document()
    section = doc.sections[0]
    section.page_width = Cm(21.0)
    section.page_height = Cm(29.7)
    section.top_margin = Cm(4.0)
    section.left_margin = Cm(4.0)
    section.bottom_margin = Cm(3.0)
    section.right_margin = Cm(3.0)
    
    style_normal = doc.styles['Normal']
    font = style_normal.font
    font.name = 'Times New Roman'
    font.size = Pt(12)
    font.color.rgb = RGBColor(0, 0, 0)
    style_normal.paragraph_format.line_spacing = 1.5
    style_normal.paragraph_format.space_after = Pt(6)
    style_normal.paragraph_format.space_before = Pt(0)
    style_normal.paragraph_format.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
    return doc

def set_cell_margins(cell, top=100, bottom=100, left=140, right=140):
    tcPr = cell._tc.get_or_add_tcPr()
    tcMar = OxmlElement('w:tcMar')
    for m, val in [('top', top), ('bottom', bottom), ('left', left), ('right', right)]:
        node = OxmlElement(f'w:{m}')
        node.set(qn('w:w'), str(val))
        node.set(qn('w:type'), 'dxa')
        tcMar.append(node)
    tcPr.append(tcMar)

def set_cell_shading(cell, color_hex):
    shading = parse_xml(f'<w:shd {nsdecls("w")} w:fill="{color_hex}"/>')
    cell._tc.get_or_add_tcPr().append(shading)

def set_academic_table_borders(table):
    tblPr = table._tbl.tblPr
    borders = parse_xml(f'''
        <w:tblBorders {nsdecls("w")}>
            <w:top w:val="single" w:sz="12" w:space="0" w:color="000000"/>
            <w:bottom w:val="single" w:sz="12" w:space="0" w:color="000000"/>
            <w:insideH w:val="single" w:sz="4" w:space="0" w:color="D3D3D3"/>
            <w:left w:val="none"/>
            <w:right w:val="none"/>
            <w:insideV w:val="none"/>
        </w:tblBorders>
    ''')
    tblPr.append(borders)

def set_no_borders(table):
    tblPr = table._tbl.tblPr
    borders = parse_xml(f'''
        <w:tblBorders {nsdecls("w")}>
            <w:top w:val="none"/>
            <w:bottom w:val="none"/>
            <w:insideH w:val="none"/>
            <w:left w:val="none"/>
            <w:right w:val="none"/>
            <w:insideV w:val="none"/>
        </w:tblBorders>
    ''')
    tblPr.append(borders)

def add_clean_paragraph(doc, text, align=WD_ALIGN_PARAGRAPH.JUSTIFY, bold=False, italic=False, space_before=0, space_after=6, line_spacing=1.5, bullet=False):
    p = doc.add_paragraph()
    p.alignment = align
    p.paragraph_format.space_before = Pt(space_before)
    p.paragraph_format.space_after = Pt(space_after)
    p.paragraph_format.line_spacing = line_spacing
    
    if bullet:
        p.paragraph_format.left_indent = Cm(0.8)
        run_b = p.add_run("•  ")
        run_b.font.name = 'Times New Roman'
        run_b.font.size = Pt(12)
        run_b.bold = True
        
    tokens = re.split(r'(\*\*[^*]+\*\*|\*[^*]+\*)', text)
    for token in tokens:
        if not token:
            continue
        if token.startswith('**') and token.endswith('**'):
            run = p.add_run(token[2:-2])
            run.font.name = 'Times New Roman'
            run.font.size = Pt(12)
            run.bold = True
        elif token.startswith('*') and token.endswith('*'):
            run = p.add_run(token[1:-1])
            run.font.name = 'Times New Roman'
            run.font.size = Pt(12)
            run.italic = True
        else:
            run = p.add_run(token)
            run.font.name = 'Times New Roman'
            run.font.size = Pt(12)
            run.bold = bold
            run.italic = italic
    return p

def add_heading_1(doc, text):
    p = doc.add_paragraph()
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p.paragraph_format.space_before = Pt(18)
    p.paragraph_format.space_after = Pt(12)
    p.paragraph_format.line_spacing = 1.15
    run = p.add_run(text.upper())
    run.font.name = 'Times New Roman'
    run.font.size = Pt(14)
    run.bold = True
    run.font.color.rgb = RGBColor(0, 0, 0)
    return p

def add_heading_2(doc, text):
    p = doc.add_paragraph()
    p.alignment = WD_ALIGN_PARAGRAPH.LEFT
    p.paragraph_format.space_before = Pt(14)
    p.paragraph_format.space_after = Pt(4)
    p.paragraph_format.line_spacing = 1.15
    run = p.add_run(text)
    run.font.name = 'Times New Roman'
    run.font.size = Pt(12)
    run.bold = True
    run.font.color.rgb = RGBColor(0, 0, 0)
    return p

def add_heading_3(doc, text):
    p = doc.add_paragraph()
    p.alignment = WD_ALIGN_PARAGRAPH.LEFT
    p.paragraph_format.space_before = Pt(10)
    p.paragraph_format.space_after = Pt(2)
    p.paragraph_format.line_spacing = 1.15
    run = p.add_run(text)
    run.font.name = 'Times New Roman'
    run.font.size = Pt(12)
    run.bold = True
    run.italic = True
    run.font.color.rgb = RGBColor(0, 0, 0)
    return p

def add_math_formula(doc, formula_text):
    p = doc.add_paragraph()
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p.paragraph_format.space_before = Pt(8)
    p.paragraph_format.space_after = Pt(8)
    p.paragraph_format.line_spacing = 1.15
    clean = formula_text.strip().replace('$$', '').replace('$', '').strip()
    run = p.add_run(clean)
    run.font.name = 'Times New Roman'
    run.font.size = Pt(11)
    run.italic = True
    return p

def add_styled_table(doc, headers, rows):
    col_count = len(headers)
    table = doc.add_table(rows=len(rows) + 1, cols=col_count)
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    table.autofit = True
    set_academic_table_borders(table)
    
    # Header
    for idx, title in enumerate(headers):
        cell = table.rows[0].cells[idx]
        cell.text = title.strip()
        set_cell_margins(cell, top=100, bottom=100, left=120, right=120)
        set_cell_shading(cell, "F0F4F8")
        p = cell.paragraphs[0]
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
        p.paragraph_format.space_before = Pt(0)
        p.paragraph_format.space_after = Pt(0)
        p.paragraph_format.line_spacing = 1.0
        for r in p.runs:
            r.font.name = 'Times New Roman'
            r.font.size = Pt(10)
            r.bold = True
            r.font.color.rgb = RGBColor(0, 32, 96)
            
    # Rows
    for r_idx, row in enumerate(rows):
        bg = "FFFFFF" if r_idx % 2 == 0 else "FBFBFC"
        for c_idx, val in enumerate(row):
            if c_idx < col_count:
                cell = table.rows[r_idx + 1].cells[c_idx]
                cell.text = val.strip()
                set_cell_margins(cell, top=80, bottom=80, left=100, right=100)
                set_cell_shading(cell, bg)
                p = cell.paragraphs[0]
                p.paragraph_format.space_before = Pt(0)
                p.paragraph_format.space_after = Pt(0)
                p.paragraph_format.line_spacing = 1.15
                v_clean = val.strip()
                if len(v_clean) <= 12 or re.match(r'^[0-9.,%+\-s~]+$', v_clean) or v_clean.startswith('S') or v_clean.startswith('CS'):
                    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
                else:
                    p.alignment = WD_ALIGN_PARAGRAPH.LEFT
                for r in p.runs:
                    r.font.name = 'Times New Roman'
                    r.font.size = Pt(9.5)
                    if "**" in v_clean or (v_clean.isupper() and len(v_clean) < 15):
                        r.bold = True
    doc.add_paragraph().paragraph_format.space_after = Pt(6)

def add_callout_box(doc, text_content):
    table = doc.add_table(rows=1, cols=1)
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    cell = table.rows[0].cells[0]
    set_cell_margins(cell, top=120, bottom=120, left=160, right=160)
    set_cell_shading(cell, "F5F7FA")
    tcPr = cell._tc.get_or_add_tcPr()
    borders = parse_xml(f'''
        <w:tcBorders {nsdecls("w")}>
            <w:left w:val="single" w:sz="24" w:space="0" w:color="003366"/>
            <w:top w:val="none"/>
            <w:right w:val="none"/>
            <w:bottom w:val="none"/>
        </w:tcBorders>
    ''')
    tcPr.append(borders)
    p = cell.paragraphs[0]
    p.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
    p.paragraph_format.space_before = Pt(0)
    p.paragraph_format.space_after = Pt(0)
    p.paragraph_format.line_spacing = 1.25
    r = p.add_run(text_content.strip())
    r.font.name = 'Times New Roman'
    r.font.size = Pt(10.5)
    r.italic = True
    doc.add_paragraph().paragraph_format.space_after = Pt(6)

def parse_markdown_to_doc(doc, md_text, is_subfile=False):
    lines = md_text.splitlines()
    i = 0
    in_mermaid = False
    in_ascii_box = False
    ascii_content = []
    
    while i < len(lines):
        line = lines[i].strip()
        
        # Skip empty lines
        if not line:
            i += 1
            continue
            
        # Mermaid code block replacement
        if line.startswith('```mermaid'):
            in_mermaid = True
            i += 1
            continue
        if in_mermaid:
            if line.startswith('```'):
                in_mermaid = False
                # Insert beautiful flow diagram table
                headers = ["Fase", "Tahapan Metodologi", "Aktivitas Kunci & Luaran"]
                rows = [
                    ["Fase 1", "Identifikasi Masalah & Studi Literatur", "Menelaah kelemahan konsultasi akademik, risiko halusinasi RAG tahap tunggal, dan teori RAG Triad."],
                    ["Fase 2", "Akuisisi Korpus & Semantic Chunking", "Pengumpulan dokumen resmi FTI UII (Pedoman UII, Skripsi FTI), kurasi, dan segmentasi 500-800 token."],
                    ["Fase 3", "Perancangan Arsitektur Pure-Go Two-Stage RAG", "Konstruksi orkestrator Go native, Dense ANN (Top-20), Neural Reranker (Top-8), dan In-Memory Parser KHS/KRS."],
                    ["Fase 4", "Pengembangan Benchmark UII-Bench-50 & Judge", "Pembuatan dataset 50 skenario (5 klaster) dan penguji otomatis LLM-as-Judge (cmd/bench)."],
                    ["Fase 5", "Pengujian Kuantitatif, Studi Ablasi & Uji Kasus", "Evaluasi 50 skenario, pengujian ablasi (+10.58%), profiling latensi komputasi, dan 15 studi kasus riil."],
                    ["Fase 6", "Analisis Pembahasan & Kesimpulan", "Sintesis hasil evaluasi RAG Triad, evaluasi kasus batas (edge cases), perumusan kesimpulan dan saran."]
                ]
                add_styled_table(doc, headers, rows)
            i += 1
            continue
            
        # ASCII architecture diagram block
        if line.startswith('```') and not in_mermaid:
            if not in_ascii_box:
                in_ascii_box = True
                ascii_content = []
                i += 1
                continue
            else:
                in_ascii_box = False
                # Insert architecture summary callout
                arch_text = "\n".join(ascii_content)
                add_callout_box(doc, arch_text)
                i += 1
                continue
                
        if in_ascii_box:
            ascii_content.append(line)
            i += 1
            continue
            
        # Markdown table detection
        if line.startswith('|') and line.endswith('|'):
            table_lines = []
            while i < len(lines) and lines[i].strip().startswith('|') and lines[i].strip().endswith('|'):
                table_lines.append(lines[i].strip())
                i += 1
            
            if len(table_lines) >= 3:
                # Header row
                raw_headers = [c.strip() for c in table_lines[0].split('|')[1:-1]]
                # Filter out delimiter row (line 1)
                data_rows = []
                for row_line in table_lines[2:]:
                    cells = [c.strip() for c in row_line.split('|')[1:-1]]
                    data_rows.append(cells)
                add_styled_table(doc, raw_headers, data_rows)
            continue
            
        # Math formula
        if line.startswith('$$') and line.endswith('$$') and len(line) > 4:
            add_math_formula(doc, line)
            i += 1
            continue
            
        # Blockquote
        if line.startswith('>'):
            quote_text = line[1:].strip()
            add_callout_box(doc, quote_text)
            i += 1
            continue
            
        # Headings
        if line.startswith('# BAB'):
            # If subfile and first line, don't page break. If master, page break before
            add_heading_1(doc, line[2:].strip())
            i += 1
            continue
        elif line.startswith('# DAFTAR'):
            add_heading_1(doc, line[2:].strip())
            i += 1
            continue
        elif line.startswith('###'):
            add_heading_3(doc, line[3:].strip())
            i += 1
            continue
        elif line.startswith('##'):
            add_heading_2(doc, line[2:].strip())
            i += 1
            continue
        elif line.startswith('#'):
            add_heading_1(doc, line[1:].strip())
            i += 1
            continue
            
        # Bullets
        if line.startswith('- ') or line.startswith('* '):
            add_clean_paragraph(doc, line[2:].strip(), bullet=True)
            i += 1
            continue
            
        # Numbered list
        m_num = re.match(r'^(\d+)\.\s+(.*)', line)
        if m_num:
            add_clean_paragraph(doc, f"{m_num.group(1)}. {m_num.group(2)}", bullet=False)
            i += 1
            continue
            
        # Regular paragraph
        if not line.startswith('---'):
            add_clean_paragraph(doc, line)
            
        i += 1

def build_all_individual_chapters():
    files = [
        ("bab1_pendahuluan", "/Users/hanifadam/AURAUII/aura-core/skripsi/bab1_pendahuluan.md"),
        ("bab2_tinjauan_pustaka", "/Users/hanifadam/AURAUII/aura-core/skripsi/bab2_tinjauan_pustaka.md"),
        ("bab3_metodologi_dan_perancangan", "/Users/hanifadam/AURAUII/aura-core/skripsi/bab3_metodologi_dan_perancangan.md"),
        ("bab4_hasil_dan_pembahasan", "/Users/hanifadam/AURAUII/aura-core/skripsi/bab4_hasil_dan_pembahasan.md"),
        ("bab5_kesimpulan_dan_saran", "/Users/hanifadam/AURAUII/aura-core/skripsi/bab5_kesimpulan_dan_saran.md"),
        ("daftar_pustaka", "/Users/hanifadam/AURAUII/aura-core/skripsi/daftar_pustaka.md")
    ]
    
    for name, path in files:
        print(f"Generating professional Word: {name}.docx...")
        doc = create_base_doc()
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
        parse_markdown_to_doc(doc, content, is_subfile=True)
        out_path = f"/Users/hanifadam/AURAUII/aura-core/skripsi/{name}.docx"
        doc.save(out_path)
        print(f"  -> Saved {out_path} ({os.path.getsize(out_path)} bytes)")

def add_cover_page(doc):
    p_title = doc.add_paragraph()
    p_title.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_title.paragraph_format.space_before = Pt(36)
    p_title.paragraph_format.space_after = Pt(12)
    p_title.paragraph_format.line_spacing = 1.15
    r = p_title.add_run("ARSITEKTUR TWO-STAGE RETRIEVAL-AUGMENTED GENERATION BERBASIS NATIVE GO UNTUK ASISTEN REGULASI AKADEMIK INSTITUSIONAL YANG TAHAN HALUSINASI\n")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(14)
    r.bold = True
    r_sub = p_title.add_run("(Studi Kasus: Fakultas Teknologi Industri Universitas Islam Indonesia)")
    r_sub.font.name = 'Times New Roman'
    r_sub.font.size = Pt(12)
    r_sub.bold = True
    
    p_skrip = doc.add_paragraph()
    p_skrip.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_skrip.paragraph_format.space_before = Pt(36)
    p_skrip.paragraph_format.space_after = Pt(6)
    r = p_skrip.add_run("SKRIPSI\n")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(14)
    r.bold = True
    r_req = p_skrip.add_run("Diajukan untuk Memenuhi Sebagian Persyaratan Mencapai Derajat Sarjana Komputer (S.Kom.)\npada Program Studi Informatika, Fakultas Teknologi Industri\nUniversitas Islam Indonesia")
    r_req.font.name = 'Times New Roman'
    r_req.font.size = Pt(11)
    
    p_author = doc.add_paragraph()
    p_author.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_author.paragraph_format.space_before = Pt(48)
    p_author.paragraph_format.space_after = Pt(36)
    p_author.paragraph_format.line_spacing = 1.2
    r_by = p_author.add_run("Disusun Oleh:\n")
    r_by.font.name = 'Times New Roman'
    r_by.font.size = Pt(11)
    r_name = p_author.add_run("MUHAMMAD NABIL HANIF\n")
    r_name.font.name = 'Times New Roman'
    r_name.font.size = Pt(12)
    r_name.bold = True
    r_nim = p_author.add_run("NIM: 23523270\n\n")
    r_nim.font.name = 'Times New Roman'
    r_nim.font.size = Pt(11)
    r_adv_lbl = p_author.add_run("Dosen Pembimbing:\n")
    r_adv_lbl.font.name = 'Times New Roman'
    r_adv_lbl.font.size = Pt(11)
    r_adv = p_author.add_run("Dr. Syarif Hidayat, S.Kom., M.I.T.")
    r_adv.font.name = 'Times New Roman'
    r_adv.font.size = Pt(12)
    r_adv.bold = True
    
    p_inst = doc.add_paragraph()
    p_inst.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_inst.paragraph_format.space_before = Pt(48)
    p_inst.paragraph_format.space_after = Pt(0)
    p_inst.paragraph_format.line_spacing = 1.15
    r_inst = p_inst.add_run("PROGRAM STUDI INFORMATIKA\nFAKULTAS TEKNOLOGI INDUSTRI\nUNIVERSITAS ISLAM INDONESIA\nYOGYAKARTA\n2026")
    r_inst.font.name = 'Times New Roman'
    r_inst.font.size = Pt(13)
    r_inst.bold = True
    doc.add_page_break()

def add_halaman_judul(doc):
    add_heading_1(doc, "HALAMAN JUDUL")
    p_title = doc.add_paragraph()
    p_title.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_title.paragraph_format.space_before = Pt(24)
    p_title.paragraph_format.space_after = Pt(12)
    p_title.paragraph_format.line_spacing = 1.15
    r = p_title.add_run("ARSITEKTUR TWO-STAGE RETRIEVAL-AUGMENTED GENERATION BERBASIS NATIVE GO UNTUK ASISTEN REGULASI AKADEMIK INSTITUSIONAL YANG TAHAN HALUSINASI\n")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.bold = True
    r_sub = p_title.add_run("(Studi Kasus: Fakultas Teknologi Industri Universitas Islam Indonesia)")
    r_sub.font.name = 'Times New Roman'
    r_sub.font.size = Pt(11)
    r_sub.bold = True

    p_skrip = doc.add_paragraph()
    p_skrip.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_skrip.paragraph_format.space_before = Pt(24)
    p_skrip.paragraph_format.space_after = Pt(6)
    r = p_skrip.add_run("SKRIPSI\n")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.bold = True
    r_req = p_skrip.add_run("Diajukan untuk Memenuhi Sebagian Persyaratan Mencapai Derajat Sarjana Komputer (S.Kom.)\npada Program Studi Informatika, Fakultas Teknologi Industri\nUniversitas Islam Indonesia")
    r_req.font.name = 'Times New Roman'
    r_req.font.size = Pt(11)

    p_author = doc.add_paragraph()
    p_author.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_author.paragraph_format.space_before = Pt(36)
    p_author.paragraph_format.space_after = Pt(24)
    p_author.paragraph_format.line_spacing = 1.2
    r_by = p_author.add_run("Disusun Oleh:\n")
    r_by.font.name = 'Times New Roman'
    r_by.font.size = Pt(11)
    r_name = p_author.add_run("Muhammad Nabil Hanif\n")
    r_name.font.name = 'Times New Roman'
    r_name.font.size = Pt(12)
    r_name.bold = True
    r_nim = p_author.add_run("NIM: 23523270\n\n")
    r_nim.font.name = 'Times New Roman'
    r_nim.font.size = Pt(11)
    r_adv_lbl = p_author.add_run("Dosen Pembimbing:\n")
    r_adv_lbl.font.name = 'Times New Roman'
    r_adv_lbl.font.size = Pt(11)
    r_adv = p_author.add_run("Dr. Syarif Hidayat, S.Kom., M.I.T.")
    r_adv.font.name = 'Times New Roman'
    r_adv.font.size = Pt(12)
    r_adv.bold = True

    p_inst = doc.add_paragraph()
    p_inst.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_inst.paragraph_format.space_before = Pt(36)
    p_inst.paragraph_format.space_after = Pt(0)
    p_inst.paragraph_format.line_spacing = 1.15
    r_inst = p_inst.add_run("PROGRAM STUDI INFORMATIKA\nFAKULTAS TEKNOLOGI INDUSTRI\nUNIVERSITAS ISLAM INDONESIA\nYOGYAKARTA\n2026")
    r_inst.font.name = 'Times New Roman'
    r_inst.font.size = Pt(12)
    r_inst.bold = True
    doc.add_page_break()

def add_pengesahan_pembimbing(doc):
    add_heading_1(doc, "HALAMAN PENGESAHAN DOSEN PEMBIMBING")
    add_clean_paragraph(doc, "Skripsi dengan judul:")
    add_clean_paragraph(doc, "**ARSITEKTUR TWO-STAGE RETRIEVAL-AUGMENTED GENERATION BERBASIS NATIVE GO UNTUK ASISTEN REGULASI AKADEMIK INSTITUSIONAL YANG TAHAN HALUSINASI**\n*(Studi Kasus: Fakultas Teknologi Industri Universitas Islam Indonesia)*", align=WD_ALIGN_PARAGRAPH.CENTER, bold=True)
    
    add_clean_paragraph(doc, "Disusun oleh:")
    add_clean_paragraph(doc, "Nama       : Muhammad Nabil Hanif\nNIM          : 23523270\nProgram Studi : Informatika\nFakultas     : Teknologi Industri\nUniversitas  : Universitas Islam Indonesia")
    
    add_clean_paragraph(doc, "Telah diperiksa, disetujui, dan disahkan oleh Dosen Pembimbing untuk diajukan dalam Ujian Pendadaran Tugas Akhir / Skripsi pada Program Studi Informatika, Fakultas Teknologi Industri, Universitas Islam Indonesia.")
    
    p_ttd = doc.add_paragraph()
    p_ttd.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    p_ttd.paragraph_format.space_before = Pt(36)
    p_ttd.paragraph_format.space_after = Pt(0)
    p_ttd.paragraph_format.line_spacing = 1.15
    r_ttd = p_ttd.add_run("Yogyakarta, September 2026\nDosen Pembimbing,\n\n\n\n\n**Dr. Syarif Hidayat, S.Kom., M.I.T.**\nNIP/NIK: 015230101")
    r_ttd.font.name = 'Times New Roman'
    r_ttd.font.size = Pt(11)
    doc.add_page_break()

def add_pengesahan_penguji(doc):
    add_heading_1(doc, "HALAMAN PENGESAHAN DOSEN PENGUJI")
    add_clean_paragraph(doc, "Skripsi dengan judul:")
    add_clean_paragraph(doc, "**ARSITEKTUR TWO-STAGE RETRIEVAL-AUGMENTED GENERATION BERBASIS NATIVE GO UNTUK ASISTEN REGULASI AKADEMIK INSTITUSIONAL YANG TAHAN HALUSINASI**\n*(Studi Kasus: Fakultas Teknologi Industri Universitas Islam Indonesia)*", align=WD_ALIGN_PARAGRAPH.CENTER, bold=True)
    
    add_clean_paragraph(doc, "Disusun oleh:")
    add_clean_paragraph(doc, "Nama       : Muhammad Nabil Hanif\nNIM          : 23523270\nProgram Studi : Informatika\nFakultas     : Teknologi Industri\nUniversitas  : Universitas Islam Indonesia")
    
    add_clean_paragraph(doc, "Telah dipertahankan di hadapan Dewan Penguji Ujian Pendadaran Tugas Akhir / Skripsi Program Studi Informatika, Fakultas Teknologi Industri, Universitas Islam Indonesia pada tanggal 21 September 2026 dan dinyatakan LULUS.")
    
    add_clean_paragraph(doc, "**Dewan Penguji:**")
    add_clean_paragraph(doc, "1. Ketua Penguji   : Dr. Syarif Hidayat, S.Kom., M.I.T.       ( ..................................... )")
    add_clean_paragraph(doc, "2. Anggota Penguji I : Hendrik, S.T., M.Eng.                     ( ..................................... )")
    add_clean_paragraph(doc, "3. Anggota Penguji II: Irving Vitra Paputungan, S.T., M.Sc., Ph.D.  ( ..................................... )")
    
    p_kps = doc.add_paragraph()
    p_kps.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_kps.paragraph_format.space_before = Pt(28)
    p_kps.paragraph_format.space_after = Pt(0)
    p_kps.paragraph_format.line_spacing = 1.15
    r_kps = p_kps.add_run("Mengetahui,\nKetua Program Studi Informatika\nFakultas Teknologi Industri\nUniversitas Islam Indonesia\n\n\n\n\n**Hendrik, S.T., M.Eng.**")
    r_kps.font.name = 'Times New Roman'
    r_kps.font.size = Pt(11)
    doc.add_page_break()

def add_pernyataan_keaslian(doc):
    add_heading_1(doc, "HALAMAN PERNYATAAN KEASLIAN")
    add_clean_paragraph(doc, "Yang bertanda tangan di bawah ini:")
    add_clean_paragraph(doc, "Nama         : Muhammad Nabil Hanif\nNIM          : 23523270\nProgram Studi : Informatika\nFakultas     : Teknologi Industri\nUniversitas  : Universitas Islam Indonesia")
    
    add_clean_paragraph(doc, "Menyatakan dengan sesungguhnya bahwa skripsi dengan judul:")
    add_clean_paragraph(doc, "**\"ARSITEKTUR TWO-STAGE RETRIEVAL-AUGMENTED GENERATION BERBASIS NATIVE GO UNTUK ASISTEN REGULASI AKADEMIK INSTITUSIONAL YANG TAHAN HALUSINASI (Studi Kasus: Fakultas Teknologi Industri Universitas Islam Indonesia)\"**", align=WD_ALIGN_PARAGRAPH.CENTER, bold=True)
    
    add_clean_paragraph(doc, "merupakan hasil karya ilmiah, pemikiran, dan penelitian orisinal saya sendiri. Sumber informasi dan rujukan yang dikutip dari karya pihak lain telah dicantumkan secara jelas dan dicantumkan dalam daftar pustaka sesuai dengan kaidah penulisan ilmiah yang berlaku.")
    
    add_clean_paragraph(doc, "Apabila di kemudian hari terbukti atau dapat dibuktikan bahwa naskah skripsi ini mengandung unsur plagiarisme atau manipulasi data, saya bersedia menerima sanksi akademik yang seberat-beratnya sesuai dengan peraturan yang berlaku di lingkungan Universitas Islam Indonesia dan ketentuan perundang-undangan Republik Indonesia.")
    
    p_ttd = doc.add_paragraph()
    p_ttd.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    p_ttd.paragraph_format.space_before = Pt(32)
    p_ttd.paragraph_format.space_after = Pt(0)
    p_ttd.paragraph_format.line_spacing = 1.15
    r_ttd = p_ttd.add_run("Yogyakarta, September 2026\nYang menyatakan,\n\n*(Meterai Rp10.000)*\n\n\n**Muhammad Nabil Hanif**\nNIM: 23523270")
    r_ttd.font.name = 'Times New Roman'
    r_ttd.font.size = Pt(11)
    doc.add_page_break()

def add_persembahan(doc):
    add_heading_1(doc, "HALAMAN PERSEMBAHAN")
    add_clean_paragraph(doc, "Dengan mengucap puji dan syukur ke hadirat Allah Subhanahu wa Ta'ala, skripsi ini saya persembahkan sebagai tanda bakti dan cinta yang tulus kepada:")
    add_clean_paragraph(doc, "Allah SWT atas segala limpahan rahmat, taufik, hidayah, kesehatan, dan keteguhan hati dalam menyelesaikan karya ilmiah ini.", bullet=True)
    add_clean_paragraph(doc, "Kedua Orang Tua Tercinta, Ayahanda dan Ibunda, yang tiada pernah lelah memanjatkan doa, mencurahkan kasih sayang tak bertepi, memberikan dukungan moril maupun materiil, serta menanamkan nilai integritas dan keteladanan sejak dini.", bullet=True)
    add_clean_paragraph(doc, "Bapak Dr. Syarif Hidayat, S.Kom., M.I.T., selaku Dosen Pembimbing, atas kesabaran, arahan konseptual, bimbingan metodologis yang presisi, serta inspirasi ilmiah yang sangat berharga sepanjang perjalanan penelitian ini.", bullet=True)
    add_clean_paragraph(doc, "Seluruh Dosen dan Tenaga Kependidikan di Lingkungan Program Studi Informatika FTI UII atas transfer ilmu pengetahuan, asistensi akademik, dan fasilitas riset yang telah diberikan selama masa studi.", bullet=True)
    add_clean_paragraph(doc, "Rekan-rekan Mahasiswa Informatika FTI UII Angkatan 2023, serta rekan sejawat di Laboratorium Riset atas kebersamaan, diskusi kritis, persahabatan, dan semangat juang yang saling menguatkan.", bullet=True)
    add_clean_paragraph(doc, "Almamater Kebanggaan, Universitas Islam Indonesia, tempat penulis menimba ilmu dan menumbuhkan komitmen keilmuan yang berlandaskan nilai-nilai keislaman serta keindonesiaan.", bullet=True)
    doc.add_page_break()

def add_moto(doc):
    add_heading_1(doc, "HALAMAN MOTO")
    motos = [
        "\"Dan katakanlah: Ya Tuhanku, tambahkanlah kepadaku ilmu pengetahuan.\"\n(QS. Thaha: 114)",
        "\"Sebaik-baik manusia di antaramu adalah yang paling banyak memberikan manfaat bagi sesama manusia.\"\n(HR. Ahmad dan Thabrani)",
        "\"Inovasi sistem cerdas berdaulat berakar pada kesederhanaan arsitektur, disiplin metodologi yang kokoh, dan komitmen tanpa henti pada integritas faktual.\"\n(Penulis)"
    ]
    for m in motos:
        add_callout_box(doc, m)
    doc.add_page_break()

def add_kata_pengantar(doc):
    add_heading_1(doc, "KATA PENGANTAR")
    p1 = ("Assalamu’alaikum Warahmatullahi Wabarakatuh,\n\n"
          "Alhamdulillahi Rabbil ‘Alamin, segala puji dan syukur penulis panjatkan ke hadirat Allah Subhanahu wa Ta’ala atas segala limpahan rahmat, hidayah, dan inayah-Nya, sehingga penulisan skripsi yang berjudul \"Arsitektur Two-Stage Retrieval-Augmented Generation Berbasis Native Go untuk Asisten Regulasi Akademik Institusional yang Tahan Halusinasi\" ini dapat terselesaikan dengan baik. Skripsi ini disusun sebagai salah satu syarat akademik guna memperoleh gelar Sarjana Komputer (S.Kom.) pada Program Studi Informatika, Fakultas Teknologi Industri, Universitas Islam Indonesia.")
    add_clean_paragraph(doc, p1)

    p2 = ("Dalam proses penelitian dan perancangan sistem AURA Core hingga penyusunan naskah ini, penulis menyadari bahwa keberhasilan karya ini tidak terlepas dari bimbingan, doa, arahan, dan bantuan dari berbagai pihak. Oleh karena itu, penulis menyampaikan rasa hormat dan terima kasih yang mendalam kepada:")
    add_clean_paragraph(doc, p2)

    ucapan = [
        "**Prof. Fathul Wahid, S.T., M.Sc., Ph.D.**, selaku Rektor Universitas Islam Indonesia.",
        "**Prof. Dr. Ir. Hari Purnomo, M.T., IPU, ASEAN.Eng.**, selaku Dekan Fakultas Teknologi Industri, Universitas Islam Indonesia.",
        "**Hendrik, S.T., M.Eng.**, selaku Ketua Program Studi Informatika, Fakultas Teknologi Industri, Universitas Islam Indonesia.",
        "**Dr. Syarif Hidayat, S.Kom., M.I.T.**, selaku Dosen Pembimbing Skripsi, atas ketelitian, waktu, arahan strategis, dan bimbingan berkesinambungan yang diberikan kepada penulis hingga penelitian ini mencapai standar mutu akademik yang diharapkan.",
        "Seluruh Dosen Program Studi Informatika FTI UII yang telah mendidik dan membagikan khazanah ilmu pengetahuan selama masa perkuliahan.",
        "Kedua orang tua dan keluarga tercinta atas limpahan doa yang tidak pernah putus, dukungan moral yang kokoh, dan pengorbanan yang tak terhingga.",
        "Rekan-rekan seperjuangan mahasiswa Informatika FTI UII Angkatan 2023 atas dinamika kebersamaan, motivasi, dan persaudaraan yang terjalin erat.",
        "Seluruh pihak yang telah membantu penyelesaian penelitian ini yang tidak dapat penulis sebutkan satu per satu."
    ]
    for idx, u in enumerate(ucapan):
        add_clean_paragraph(doc, f"{idx+1}. {u}")

    p3 = ("Penulis menyadari bahwa skripsi ini masih memiliki keterbatasan. Masukan, kritik, dan saran konstruktif dari pembaca sangat diharapkan demi perbaikan di masa mendatang. Akhir kata, semoga naskah skripsi ini bermanfaat bagi kemajuan ilmu pengetahuan, sivitas akademika Universitas Islam Indonesia, dan masyarakat luas.\n\n"
          "Wassalamu’alaikum Warahmatullahi Wabarakatuh.")
    add_clean_paragraph(doc, p3)

    p_ttd = doc.add_paragraph()
    p_ttd.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    p_ttd.paragraph_format.space_before = Pt(24)
    p_ttd.paragraph_format.space_after = Pt(0)
    p_ttd.paragraph_format.line_spacing = 1.15
    r_ttd = p_ttd.add_run("Yogyakarta, September 2026\nPenulis,\n\n\n**Muhammad Nabil Hanif**\nNIM: 23523270")
    r_ttd.font.name = 'Times New Roman'
    r_ttd.font.size = Pt(11)
    doc.add_page_break()

def add_sari(doc):
    add_heading_1(doc, "SARI")
    sari_text = (
        "Konsultasi regulasi akademik di perguruan tinggi menuntut akurasi faktual mutlak, toleransi halusinasi nol, "
        "serta perlindungan ketat terhadap data pribadi mahasiswa. Implementasi Retrieval-Augmented Generation (RAG) tahap tunggal "
        "konvensional yang mengandalkan kemiripan vektor tunggal berdimensi tinggi sering kali mengalami pergeseran semantik (semantic drift), "
        "tertukarnya klausul hukum, serta menghasilkan izin regulasi fiktif (halusinasi). Selain itu, ketergantungan pada kerangka kerja Python "
        "yang berat menimbulkan overhead komputasi tinggi dan kerentanan kebocoran data akademik jika transkrip mahasiswa disimpan ke basis data vektor publik.\n\n"
        "Penelitian ini merancang dan mengimplementasikan AURA Core, sebuah sistem asisten regulasi akademik berbasis arsitektur Pure-Go Native Two-Stage RAG "
        "yang dirancang secara berdaulat untuk Fakultas Teknologi Industri, Universitas Islam Indonesia (FTI UII). Sistem mengintegrasikan penelusuran "
        "vektor awal Dense ANN Retrieval (Top-20 kandidat) menggunakan representasi multibahasa 1024-dimensi yang disaring kembali oleh Cross-Encoder Neural Reranker "
        "(Top-8 klausul terpilih), dipadukan dengan penalaran LLM bersuhu rendah (tau = 0.2) serta kewajiban sitasi nomor pasal resmi. Privasi data mahasiswa "
        "dilindungi secara deterministik melalui in-memory parser transkrip KHS/KRS di RAM per sesi transaksi tanpa persistensi ke repositori vektor publik.\n\n"
        "Evaluasi empiris dilakukan menggunakan dataset tolok ukur institusional UII-Bench-50 (50 skenario dalam 5 klaster regulasi) dan 15 studi kasus "
        "siklus hidup mahasiswa melalui pengujian otomatis LLM-as-Judge RAG Triad. Hasil pengujian menunjukkan AURA Core mencapai rata-rata Context Relevance "
        "sebesar 87,40%, Groundedness (Anti-Halusinasi) sebesar 91,80%, Answer Relevance sebesar 98,60%, dan RAG Triad Score sebesar 90,55% dengan latensi median "
        "end-to-end 4,24 detik. Studi ablasi membuktikan bahwa penambahan tahap Neural Reranking menghasilkan lonjakan faktualitas Groundedness sebesar +10,58% "
        "(dari 81,22% menjadi 91,80%) dengan penambahan latensi hanya 330 ms, membuktikan keunggulan arsitektur Two-Stage RAG untuk tata kelola akademik institusi."
    )
    for p_seg in sari_text.split('\n\n'):
        add_clean_paragraph(doc, p_seg, line_spacing=1.15, space_after=6)
        
    p_kw = doc.add_paragraph()
    p_kw.paragraph_format.space_before = Pt(8)
    p_kw.paragraph_format.space_after = Pt(12)
    p_kw.paragraph_format.line_spacing = 1.15
    r_kwh = p_kw.add_run("Kata Kunci: ")
    r_kwh.bold = True
    r_kwt = p_kw.add_run("Retrieval-Augmented Generation, Regulasi Akademik, Neural Reranking, Anti-Halusinasi, Pure Go, UII-Bench-50, FTI UII.")
    r_kwt.italic = True
    doc.add_page_break()

def add_abstract(doc):
    add_heading_1(doc, "ABSTRACT")
    abstract_text = (
        "Advising university students on institutional academic regulations requires absolute factual fidelity, zero hallucination tolerance, "
        "and strict data privacy. Conventional single-stage Retrieval-Augmented Generation (RAG) implementations typically rely on single-stage vector "
        "similarity searches wrapped in heavy runtime frameworks, which frequently introduce semantic drift, clause confusion, and hallucinated regulatory allowances. "
        "Furthermore, personalized student advising requires parsing sensitive transcripts, posing severe data leakage risks if ingested into communal vector indices.\n\n"
        "This study designs and implements AURA Core, a sovereign, high-performance Pure-Go native Two-Stage RAG architecture tailored specifically "
        "for the Faculty of Industrial Technology at Universitas Islam Indonesia (FTI UII). The system cascades a first-stage Dense Approximate Nearest Neighbor (ANN) "
        "vector retrieval (top-20 candidates using 1024-dimensional embeddings) with a second-stage Cross-Encoder Neural Reranker (top-8 candidates) and low-temperature "
        "reasoning (tau = 0.2) with mandatory article-level citation constraints. Student academic transcripts (KHS/KRS) are parsed strictly in-memory per session, "
        "guaranteeing zero persistence in public vector stores.\n\n"
        "We rigorously evaluate the system using UII-Bench-50 (50 institutional scenarios across five regulatory clusters) and 15 real-world student lifecycle cases "
        "scored via automated LLM-as-Judge RAG Triad metrics. Experimental results demonstrate that AURA Core achieves an overall Context Relevance of 87.40%, "
        "Groundedness of 91.80%, Answer Relevance of 98.60%, and a composite Triad Score of 90.55% with a median end-to-end latency of 4.24 seconds. "
        "A controlled ablation study proves that Stage-2 Neural Reranking is vital, conferring a +10.58% leap in Groundedness over the single-stage baseline "
        "(81.22% vs. 91.80%) at an incremental latency overhead of only 330 ms."
    )
    for p_seg in abstract_text.split('\n\n'):
        p_eng = doc.add_paragraph()
        p_eng.paragraph_format.line_spacing = 1.15
        p_eng.paragraph_format.space_after = Pt(6)
        p_eng.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
        r_eng = p_eng.add_run(p_seg)
        r_eng.font.name = 'Times New Roman'
        r_eng.font.size = Pt(11)
        r_eng.italic = True
        
    p_kw_eng = doc.add_paragraph()
    p_kw_eng.paragraph_format.space_before = Pt(8)
    p_kw_eng.paragraph_format.space_after = Pt(12)
    p_kw_eng.paragraph_format.line_spacing = 1.15
    r_kwh_e = p_kw_eng.add_run("Keywords: ")
    r_kwh_e.bold = True
    r_kwt_e = p_kw_eng.add_run("Retrieval-Augmented Generation, Institutional Knowledge Base, Neural Reranking, Anti-Hallucination, Academic Governance, Pure Go.")
    r_kwt_e.italic = True
    doc.add_page_break()

def add_glosarium(doc):
    add_heading_1(doc, "GLOSARIUM")
    headers = ["Istilah", "Keterangan Definisi"]
    rows = [
        ["ANN (Approximate Nearest Neighbor)", "Algoritma pencarian vektor tetangga terdekat berbasis kemiripan kosinus berdimensi tinggi untuk penelusuran cepat pada ruang semantik besar."],
        ["Bi-Encoder", "Arsitektur model transformator yang memetakan kueri dan dokumen ke dalam vektor terpisah secara independen."],
        ["Context Relevance", "Rasio kesesuaian dan kebergunaan klausul regulasi yang diambil oleh retrieval terhadap kebutuhan pertanyaan pengguna."],
        ["Cross-Encoder", "Arsitektur neural network yang memproses pasangan kueri dan kandidat dokumen secara bersamaan guna menghasilkan skor relevansi semantik tingkat tinggi."],
        ["Groundedness", "Metrik anti-halusinasi yang mengukur sejauh mana setiap klaim dalam respons sistem didukung secara faktual oleh klausul regulasi rujukan."],
        ["In-Memory Parsing", "Ekstraksi dan evaluasi data transkrip akademik mahasiswa secara langsung di memori kerja (RAM) tanpa persistensi ke basis data permanen."],
        ["IPS (Indeks Prestasi Semester)", "Nilai rata-rata capaian belajar mahasiswa pada satu semester yang menentukan kuota pengambilan SKS semester berikutnya."],
        ["KHS (Kartu Hasil Studi)", "Dokumen transkrip resmi per semester yang memuat daftar nilai capaian mata kuliah mahasiswa."],
        ["KRS (Kartu Rencana Studi)", "Formulir rencana pengambilan mata kuliah yang diajukan mahasiswa pada awal semester perkuliahan."],
        ["LLM (Large Language Model)", "Model pembelajaran mesin berbasis deep learning berskala parameter miliaran yang dilatih untuk memahami dan menghasilkan bahasa alami."],
        ["Lost-in-the-Middle", "Fenomena degradasi atensi LLM ketika informasi penting berada pada bagian tengah jendela konteks yang panjang."],
        ["Pure Go", "Implementasi sistem perangkat lunak yang dibangun murni menggunakan bahasa Go tanpa runtime Python eksternal maupun wrapper CGO."],
        ["RAG (Retrieval-Augmented Generation)", "Metode pengayaan konteks penalaran LLM menggunakan dokumen rujukan yang relevan dari korpus pengetahuan eksternal."],
        ["RAG Triad", "Kerangka kerja tiga serangkai metrik evaluasi RAG yang mencakup Context Relevance, Groundedness, dan Answer Relevance."],
        ["Neural Reranking", "Komputasi pemeringkatan ulang tahap kedua menggunakan model neural untuk menyaring kandidat dokumen paling relevan sebelum diserahkan ke LLM."],
        ["Semantic Drift", "Pergeseran makna semantik ketika pencarian kemiripan vektor mengambil pasal yang serupa secara kata namun salah secara konteks hukum."],
        ["SK Pembimbing", "Surat Keputusan pimpinan fakultas mengenai penugasan dosen pembimbing tugas akhir/skripsi mahasiswa."],
        ["Turnitin", "Sistem komputasi institusional untuk memeriksa tingkat kemiripan tekstual guna mencegah plagiarisme pada naskah ilmiah."]
    ]
    add_styled_table(doc, headers, rows)
    doc.add_page_break()

def add_leader_table(doc, items):
    t = doc.add_table(rows=len(items), cols=2)
    t.alignment = WD_TABLE_ALIGNMENT.CENTER
    t.autofit = False
    set_no_borders(t)
    for r_i, (title, page) in enumerate(items):
        c0 = t.rows[r_i].cells[0]
        c1 = t.rows[r_i].cells[1]
        c0.width = Cm(12.5)
        c1.width = Cm(1.5)
        set_cell_margins(c0, top=30, bottom=30, left=40, right=40)
        set_cell_margins(c1, top=30, bottom=30, left=40, right=40)
        
        p0 = c0.paragraphs[0]
        p0.paragraph_format.space_before = Pt(0)
        p0.paragraph_format.space_after = Pt(0)
        p0.paragraph_format.line_spacing = 1.15
        r0 = p0.add_run(title)
        r0.font.name = 'Times New Roman'
        r0.font.size = Pt(11)
        if title.startswith("BAB") or title in ["HALAMAN JUDUL", "HALAMAN PENGESAHAN DOSEN PEMBIMBING", "HALAMAN PENGESAHAN DOSEN PENGUJI", "HALAMAN PERNYATAAN KEASLIAN", "HALAMAN PERSEMBAHAN", "HALAMAN MOTO", "KATA PENGANTAR", "SARI", "ABSTRACT", "GLOSARIUM", "DAFTAR ISI", "DAFTAR TABEL", "DAFTAR GAMBAR", "DAFTAR PUSTAKA"]:
            r0.bold = True
            
        p1 = c1.paragraphs[0]
        p1.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        p1.paragraph_format.space_before = Pt(0)
        p1.paragraph_format.space_after = Pt(0)
        p1.paragraph_format.line_spacing = 1.15
        r1 = p1.add_run(page)
        r1.font.name = 'Times New Roman'
        r1.font.size = Pt(11)
        if title.startswith("BAB") or title in ["HALAMAN JUDUL", "HALAMAN PENGESAHAN DOSEN PEMBIMBING", "HALAMAN PENGESAHAN DOSEN PENGUJI", "HALAMAN PERNYATAAN KEASLIAN", "HALAMAN PERSEMBAHAN", "HALAMAN MOTO", "KATA PENGANTAR", "SARI", "ABSTRACT", "GLOSARIUM", "DAFTAR ISI", "DAFTAR TABEL", "DAFTAR GAMBAR", "DAFTAR PUSTAKA"]:
            r1.bold = True
    doc.add_page_break()

def add_daftar_isi(doc):
    add_heading_1(doc, "DAFTAR ISI")
    toc_items = [
        ("HALAMAN JUDUL", "i"),
        ("HALAMAN PENGESAHAN DOSEN PEMBIMBING", "ii"),
        ("HALAMAN PENGESAHAN DOSEN PENGUJI", "iii"),
        ("HALAMAN PERNYATAAN KEASLIAN", "iv"),
        ("HALAMAN PERSEMBAHAN", "v"),
        ("HALAMAN MOTO", "vi"),
        ("KATA PENGANTAR", "vii"),
        ("SARI", "ix"),
        ("ABSTRACT", "x"),
        ("GLOSARIUM", "xi"),
        ("DAFTAR ISI", "xii"),
        ("DAFTAR TABEL", "xiv"),
        ("DAFTAR GAMBAR", "xv"),
        ("BAB I PENDAHULUAN", "1"),
        ("  1.1 Latar Belakang Masalah", "1"),
        ("  1.2 Rumusan Masalah", "4"),
        ("  1.3 Batasan Masalah", "4"),
        ("  1.4 Tujuan Penelitian", "5"),
        ("  1.5 Manfaat Penelitian", "5"),
        ("  1.6 Sistematika Penulisan", "6"),
        ("BAB II TINJAUAN PUSTAKA DAN LANDASAN TEORI", "7"),
        ("  2.1 Tinjauan Pustaka (State of the Art)", "7"),
        ("  2.2 Landasan Teori", "9"),
        ("    2.2.1 Large Language Models (LLM) dan Fenomena Halusinasi", "9"),
        ("    2.2.2 Konsep Dasar Retrieval-Augmented Generation (RAG)", "10"),
        ("    2.2.3 Representasi Vektor Semantik dan Pencarian ANN", "11"),
        ("    2.2.4 Cross-Encoder Neural Reranking", "12"),
        ("    2.2.5 Fenomena Lost-in-the-Middle", "13"),
        ("    2.2.6 Kerangka Evaluasi RAG Triad Berbasis LLM-as-Judge", "14"),
        ("    2.2.7 Karakteristik Bahasa Pemrograman Go", "15"),
        ("BAB III METODOLOGI PENELITIAN DAN PERANCANGAN SISTEM", "16"),
        ("  3.1 Alur Metodologi Penelitian", "16"),
        ("  3.2 Akuisisi dan Prapemrosesan Korpus Regulasi FTI UII", "17"),
        ("  3.3 Perancangan Arsitektur Sistem AURA Core (Pure-Go Native)", "18"),
        ("    3.3.1 Tahap 1: Dense Vector Retrieval (Bi-Encoder ANN)", "19"),
        ("    3.3.2 Tahap 2: Cross-Encoder Neural Reranking", "20"),
        ("    3.3.3 Mekanisme Privasi Mahasiswa (In-Memory Privacy Parser)", "21"),
        ("    3.3.4 Tahap 3: Constrained Grounded Reasoning", "22"),
        ("  3.4 Desain Pengujian dan Metrik Evaluasi", "23"),
        ("    3.4.1 Dataset Tolok Ukur UII-Bench-50", "23"),
        ("    3.4.2 Implementasi Automated LLM-as-Judge Runner", "24"),
        ("    3.4.3 Desain Eksperimen Studi Ablasi", "25"),
        ("    3.4.4 Desain 15 Studi Kasus Siklus Hidup Mahasiswa", "26"),
        ("BAB IV HASIL DAN PEMBAHASAN", "27"),
        ("  4.1 Lingkungan Implementasi dan Pengujian", "27"),
        ("  4.2 Hasil Evaluasi Kuantitatif UII-Bench-50", "28"),
        ("  4.3 Pembahasan Studi Ablasi: Single-Stage vs Two-Stage RAG", "32"),
        ("  4.4 Analisis Profil Latensi dan Efisiensi Komputasi Pipeline Pure-Go", "34"),
        ("  4.5 Hasil Pengujian Kualitatif: 15 Studi Kasus Mahasiswa", "35"),
        ("  4.6 Pembahasan Mendalam Kasus Kritis dan Ketahanan Halusinasi", "38"),
        ("BAB V KESIMPULAN DAN SARAN", "41"),
        ("  5.1 Kesimpulan", "41"),
        ("  5.2 Saran", "42"),
        ("DAFTAR PUSTAKA", "43")
    ]
    add_leader_table(doc, toc_items)

def add_daftar_tabel(doc):
    add_heading_1(doc, "DAFTAR TABEL")
    table_items = [
        ("Tabel 1.1 Pemetaan Masalah Konsultasi Regulasi Akademik vs. Solusi AURA Core", "3"),
        ("Tabel 2.1 Perbandingan State-of-the-Art Pendekatan RAG pada Regulasi Akademik", "8"),
        ("Tabel 3.1 Karakteristik dan Distribusi Korpus Regulasi Akademik FTI UII", "17"),
        ("Tabel 3.2 Distribusi 50 Skenario Uji UII-Bench-50 Berdasarkan Klaster Regulasi", "24"),
        ("Tabel 3.3 Karakteristik 15 Profil Mahasiswa pada Pengujian Studi Kasus Riil", "26"),
        ("Tabel 4.1 Hasil Evaluasi Kuantitatif UII-Bench-50 pada Arsitektur AURA Core", "28"),
        ("Tabel 4.2 Hasil Evaluasi Studi Kasus Riil Mahasiswa (CS01 - CS15)", "36"),
        ("Tabel 4.3 Perbandingan Kinerja Studi Ablasi (Single-Stage vs. Two-Stage RAG)", "33"),
        ("Tabel 4.4 Analisis Perbandingan Profil Latensi dan Overhead Pipeline Pure-Go", "34")
    ]
    add_leader_table(doc, table_items)

def add_daftar_gambar(doc):
    add_heading_1(doc, "DAFTAR GAMBAR")
    figure_items = [
        ("Gambar 3.1 Alur Metodologi Penelitian Pengembangan Sistem AURA Core", "16"),
        ("Gambar 3.2 Arsitektur Kaskade Two-Stage RAG Berbasis Pure-Go Native", "19"),
        ("Gambar 3.3 Diagram Alir Pengujian Otomatis Menggunakan LLM-as-Judge Runner", "24"),
        ("Gambar 4.1 Grafik Perbandingan Triad Score Antar-Klaster Regulasi Akademik", "29"),
        ("Gambar 4.2 Profil Distribusi Latensi Komputasi Pipeline AURA Core Pure-Go", "35")
    ]
    add_leader_table(doc, figure_items)

def build_master_skripsi():
    print("Generating Master Unified Skripsi Document: skripsi_lengkap_aura_core.docx...")
    doc = create_base_doc()
    
    # 1. Cover Page
    add_cover_page(doc)
    # 2. Halaman Judul
    add_halaman_judul(doc)
    # 3. Halaman Pengesahan Dosen Pembimbing
    add_pengesahan_pembimbing(doc)
    # 4. Halaman Pengesahan Dosen Penguji
    add_pengesahan_penguji(doc)
    # 5. Halaman Pernyataan Keaslian
    add_pernyataan_keaslian(doc)
    # 6. Halaman Persembahan
    add_persembahan(doc)
    # 7. Halaman Moto
    add_moto(doc)
    # 8. Kata Pengantar
    add_kata_pengantar(doc)
    # 9. Sari (Abstrak Bahasa Indonesia)
    add_sari(doc)
    # 10. Abstract (English)
    add_abstract(doc)
    # 11. Glosarium
    add_glosarium(doc)
    # 12. Daftar Isi
    add_daftar_isi(doc)
    # 13. Daftar Tabel
    add_daftar_tabel(doc)
    # 14. Daftar Gambar
    add_daftar_gambar(doc)
    
    # 15. Chapters & References
    chapter_paths = [
        "/Users/hanifadam/AURAUII/aura-core/skripsi/bab1_pendahuluan.md",
        "/Users/hanifadam/AURAUII/aura-core/skripsi/bab2_tinjauan_pustaka.md",
        "/Users/hanifadam/AURAUII/aura-core/skripsi/bab3_metodologi_dan_perancangan.md",
        "/Users/hanifadam/AURAUII/aura-core/skripsi/bab4_hasil_dan_pembahasan.md",
        "/Users/hanifadam/AURAUII/aura-core/skripsi/bab5_kesimpulan_dan_saran.md",
        "/Users/hanifadam/AURAUII/aura-core/skripsi/daftar_pustaka.md"
    ]
    
    for idx, path in enumerate(chapter_paths):
        print(f"Adding chapter {idx+1}...")
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
        parse_markdown_to_doc(doc, content, is_subfile=False)
        if idx < len(chapter_paths) - 1:
            doc.add_page_break()
            
    out_path = "/Users/hanifadam/AURAUII/aura-core/skripsi/skripsi_lengkap_aura_core.docx"
    doc.save(out_path)
    print(f"Master Skripsi generated: {out_path} ({os.path.getsize(out_path)} bytes)")
    
    # Also save as Skripsi_Final_Muhammad Nabil Hanif.docx
    final_path = "/Users/hanifadam/AURAUII/aura-core/skripsi/Skripsi_Final_Muhammad Nabil Hanif.docx"
    doc.save(final_path)
    print(f"Final Skripsi saved: {final_path} ({os.path.getsize(final_path)} bytes)")

if __name__ == "__main__":
    build_all_individual_chapters()
    build_master_skripsi()
    print("ALL WORD DOCUMENTS GENERATED WITH IMPECCABLE QUALITY!")
