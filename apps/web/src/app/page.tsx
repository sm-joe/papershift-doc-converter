"use client";

import { useEffect, useRef, useState } from "react";

type Format = {
  id: string;
  name: string;
  extension: string;
  mime_types: string[];
  family: string;
};

type Capability = {
  input: Format;
  output: Format;
  engine: string;
};

type CapabilitiesResponse = {
  capabilities: Capability[];
};

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function extensionOf(filename: string) {
  const index = filename.lastIndexOf(".");

  if (index === -1) {
    return "";
  }

  return filename.slice(index + 1).toLowerCase();
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function Home() {
  const inputRef = useRef<HTMLInputElement>(null);
  const mergeInputRef = useRef<HTMLInputElement>(null);

  const [file, setFile] = useState<File | null>(null);
  const [dragging, setDragging] = useState(false);
  const [capabilities, setCapabilities] = useState<Capability[]>([]);
  const [loadingCapabilities, setLoadingCapabilities] = useState(true);
  const [capabilitiesError, setCapabilitiesError] = useState("");
  const [outputFormat, setOutputFormat] = useState("");
  const [converting, setConverting] = useState(false);
  const [conversionError, setConversionError] = useState("");
  const [downloadUrl, setDownloadUrl] = useState("");
  const [mergeMode, setMergeMode] = useState(false);
  const [mergeFiles, setMergeFiles] = useState<File[]>([]);
  const [merging, setMerging] = useState(false);
  const [rotateMode, setRotateMode] = useState(false);
  const [rotation, setRotation] = useState(90);
  const [rotating, setRotating] = useState(false);

  useEffect(() => {
    async function loadCapabilities() {
      try {
        const response = await fetch(
          `${API_URL}/api/v1/capabilities`,
        );

        if (!response.ok) {
          throw new Error("Failed to load conversion capabilities");
        }

        const data: CapabilitiesResponse = await response.json();

        setCapabilities(data.capabilities);
      } catch (error) {
        console.error(error);
        
        setCapabilitiesError(
          "Conversion formats could not be loaded. Please make sure the API is running.",
        );
      } finally {
        setLoadingCapabilities(false);
      }
    }

    loadCapabilities();
  }, []);

  function selectFile(selectedFile: File | undefined) {
    if (!selectedFile) {
      return;
    }

    setFile(selectedFile);
    setOutputFormat("");
    setConversionError("");
    setDownloadUrl("");
  }

  function selectMergeFiles(selectedFiles: FileList | null) {
  if (!selectedFiles) {
    return;
  }

  const pdfFiles = Array.from(selectedFiles).filter(
    (selectedFile) =>
      selectedFile.type === "application/pdf" ||
      extensionOf(selectedFile.name) === "pdf",
  );

  setMergeFiles(pdfFiles);
  setConversionError("");
  setDownloadUrl("");
}

  function handleDrop(event: React.DragEvent<HTMLDivElement>) {
    event.preventDefault();
    setDragging(false);

    selectFile(event.dataTransfer.files?.[0]);
  }

  function resetWorkspace() {
  setFile(null);
  setOutputFormat("");
  setConversionError("");
  setDownloadUrl("");

  if (inputRef.current) {
    inputRef.current.value = "";
  }
} 

  async function convertFile() {
    if (!file || !outputFormat) {
      return;
    }

    setConverting(true);
    setOutputFormat(outputFormat);
    setConversionError("");
    setDownloadUrl("");

    try {
      const formData = new FormData();

      formData.append("file", file);
      formData.append("output_format", outputFormat);

      const response = await fetch(
        `${API_URL}/api/v1/conversions`,
        {
          method: "POST",
          body: formData,
        },
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data.error ?? "Conversion failed",
        );
      }

      setDownloadUrl(`${API_URL}${data.output}`);
    } catch (error) {
      setConversionError(
        error instanceof Error
          ? "PaperShift could not reach the conversion service. Please make sure the API is running."
          : error instanceof Error
          ? error.message
          : "Conversion failed. Please try again.",
      );
    } finally {
      setConverting(false);
    }
  }

  async function mergePDFs() {
  if (mergeFiles.length < 2) {
    return;
  }

  setMerging(true);
  setConversionError("");
  setDownloadUrl("");

  try {
    const formData = new FormData();

    mergeFiles.forEach((mergeFile) => {
      formData.append("files", mergeFile);
    });

    const response = await fetch(
      `${API_URL}/api/v1/pdf/merge`,
      {
        method: "POST",
        body: formData,
      },
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error ?? "PDF merge failed");
    }

    setDownloadUrl(`${API_URL}${data.output}`);
  } catch (error) {
    setConversionError(
      error instanceof Error
        ? error.message
        : "PDF merge failed. Please try again.",
    );
  } finally {
    setMerging(false);
  }
}

async function rotatePDF() {
  if (!file) {
    return;
  }

  setRotating(true);
  setConversionError("");
  setDownloadUrl("");

  try {
    const formData = new FormData();

    formData.append("file", file);
    formData.append("rotation", String(rotation));

    const response = await fetch(
      `${API_URL}/api/v1/pdf/rotate`,
      {
        method: "POST",
        body: formData,
      },
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error ?? "PDF rotation failed");
    }

    setDownloadUrl(`${API_URL}${data.output}`);
  } catch (error) {
    setConversionError(
      error instanceof Error
        ? error.message
        : "PDF rotation failed. Please try again.",
    );
  } finally {
    setRotating(false);
  }
}

  const inputFormatID = file
    ? extensionOf(file.name)
    : "";

  const outputFormats = capabilities
    .filter(
      (capability) =>
        capability.input.id === inputFormatID,
    )
    .map((capability) => capability.output)
    .filter(
      (format, index, formats) =>
        formats.findIndex(
          (item) => item.id === format.id,
        ) === index,
    );

  const detectedFormat = capabilities.find(
    (capability) =>
      capability.input.id === inputFormatID,
  )?.input;

  const isSupportedInput = capabilities.some(
  (capability) =>
    capability.input.id === inputFormatID,
  );

  return (
    <main className="min-h-screen overflow-hidden bg-[#f4f1ea] text-[#171717]">
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute -right-32 -top-32 h-96 w-96 rounded-full bg-[#d8e4dc] blur-3xl" />
        <div className="absolute -bottom-40 -left-32 h-96 w-96 rounded-full bg-[#e8d8c5] blur-3xl" />
      </div>

      <div className="relative mx-auto max-w-7xl px-5 py-5 sm:px-8 lg:px-10">
        <header className="flex items-center justify-between border-b border-black/10 pb-5">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 rotate-[-6deg] items-center justify-center rounded-lg bg-[#171717] text-sm font-bold text-white shadow-sm">
              P
            </div>

            <div>
              <div className="text-[17px] font-semibold tracking-[-0.03em]">
                PaperShift
              </div>

              <div className="text-[10px] uppercase tracking-[0.18em] text-black/40">
                File transformation studio
              </div>
            </div>
          </div>

          <div className="hidden items-center gap-2 sm:flex">
            <span className="rounded-full border border-black/10 bg-white/50 px-3 py-1.5 text-xs text-black/55">
              Open Source
            </span>

            <span className="rounded-full bg-[#dce9df] px-3 py-1.5 text-xs font-medium text-[#31513d]">
              Self-Hosted
            </span>
          </div>
        </header>

        <section className="grid min-h-[calc(100vh-110px)] items-center gap-12 py-12 lg:grid-cols-[0.9fr_1.1fr] lg:gap-20">
          <div className="max-w-xl">
            <div className="mb-7 flex items-center gap-3">
              <span className="h-px w-8 bg-black/30" />

              <span className="text-[11px] font-semibold uppercase tracking-[0.22em] text-black/45">
                Convert without the clutter
              </span>
            </div>

            <h1 className="text-[clamp(3.5rem,7vw,6.8rem)] font-semibold leading-[0.84] tracking-[-0.075em]">
              Move
              <br />
              files.
              <br />
              <span className="font-serif font-normal italic text-[#777268]">
                Not headaches.
              </span>
            </h1>

            <p className="mt-8 max-w-md text-[15px] leading-7 text-black/55">
              A focused workspace for turning documents,
              spreadsheets, presentations, and images into
              the formats you actually need.
            </p>

            <div className="mt-9 flex flex-wrap gap-2">
              {[
                "Documents",
                "Spreadsheets",
                "Presentations",
                "Images",
              ].map((item) => (
                <span
                  key={item}
                  className="rounded-full border border-black/10 bg-white/45 px-3.5 py-2 text-xs text-black/55"
                >
                  {item}
                </span>
              ))}
            </div>
          </div>

          <div className="relative">
            <div className="absolute -right-3 -top-3 h-full w-full rotate-[2deg] rounded-[28px] border border-black/10 bg-[#e8e3d9]" />

            <div className="relative rounded-[28px] border border-black/10 bg-[#fffdfa] p-4 shadow-[0_30px_80px_rgba(40,35,25,0.12)] sm:p-6">
              <div className="mb-5 flex items-center justify-between px-1">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.16em] text-black/35">
                    Conversion desk
                  </p>

                  <div className="mt-1 flex items-center gap-2 text-sm">
                    <button
                      type="button"
                      onClick={() => {
                        setMergeMode(false);
                        setRotateMode(false);
                        setMergeFiles([]);
                        setConversionError("");
                        setDownloadUrl("");
                      }}
                      className={
                        !mergeMode
                        ? "rounded-lg bg-black/5 px-3 py-1.5 font-medium text-black"
                        : "rounded-lg px-3 py-1.5 text-black/40 transition hover:bg-black/5 hover:text-black/70"
                      }
                    >
                      Convert Files
                    </button>

                    <span className="text-black/20">·</span>
                    <button
                      type="button"
                      onClick={() => {
                        setMergeMode(true);
                        setRotateMode(false);
                        setFile(null);
                        setOutputFormat("");
                        setConversionError("");
                        setDownloadUrl("");
                        setTimeout(() => mergeInputRef.current?.click(), 0);
                      }}
                      className={
                        mergeMode
                        ? "rounded-lg bg-[#dce9df] px-3 py-1.5 font-medium text-[#31513d]"
                        : "rounded-lg px-3 py-1.5 text-black/40 transition hover:bg-black/5 hover:text-black/70"
                      }
                    >
                      Merge PDFs
                    </button>
                    
                    <span className="text-black/20">·</span>
                    <button
                      type="button"
                      onClick={() => {
                        setMergeMode(false);
                        setRotateMode(true);
                        setMergeFiles([]);
                        setFile(null);
                        setOutputFormat("");
                        setConversionError("");
                        setDownloadUrl("");
                        setRotation(90);
                        setTimeout(() => inputRef.current?.click(), 0);
                      }}
                      className={
                        rotateMode
                          ? "rounded-lg bg-[#dce9df] px-3 py-1.5 font-medium text-[#31513d]"
                          : "rounded-lg px-3 py-1.5 text-black/40 transition hover:bg-black/5 hover:text-black/70"
                      }
                    >
                      Rotate PDF
                    </button>


                  </div>
                </div>
                </div>

                <div className="flex gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-[#d8c9b4]" />
                  <span className="h-2 w-2 rounded-full bg-[#b8cbbd]" />
                  <span className="h-2 w-2 rounded-full bg-[#c9c2b8]" />
                </div>
              </div>

              <div
                onDragOver={(event) => {
                  event.preventDefault();
                  setDragging(true);
                }}
                onDragLeave={() => setDragging(false)}
                onDrop={handleDrop}
                onClick={() => {
                  if (!rotateMode) {
                    inputRef.current?.click();
                  }
                }}
                className={`group relative cursor-pointer overflow-hidden rounded-[22px] border transition ${
                  dragging
                    ? "border-[#557361] bg-[#edf5ef]"
                    : file
                      ? "border-[#b9cbbd] bg-[#f2f6f1]"
                      : "border-black/10 bg-[#f7f5ef] hover:border-black/20 hover:bg-[#f3f0e8]"
                }`}
              >
                {mergeMode ? (
                  <input
                  ref={mergeInputRef}
                  type="file"
                  className="hidden"
                  multiple
                  accept="application/pdf,.pdf"
                  onChange={(event) =>
                    selectMergeFiles(event.target.files)
                  }
                />
                ) : (

                  <input

                    ref={inputRef}

                    type="file"

                    className="hidden"

                    accept={rotateMode ? "application/pdf,.pdf" : undefined}

                    onChange={(event) => {

                      const selectedFile = event.target.files?.[0];


                      if (rotateMode) {

                        if (

                          selectedFile &&

                          (

                            selectedFile.type === "application/pdf" ||

                            extensionOf(selectedFile.name) === "pdf"

                          )

                        ) {

                          selectFile(selectedFile);

                        } else if (selectedFile) {

                          setFile(null);

                          setConversionError("Rotate PDF only supports PDF files.");

                        }

                      } else {

                        selectFile(selectedFile);

                      }

                    }}

                  />

                )}

                <div className="relative flex min-h-[270px] flex-col items-center justify-center px-6 py-12 text-center">
                  <div
                    className={`relative mb-7 h-20 w-16 rounded-md border border-black/15 bg-white shadow-[6px_8px_0_#ded9ce] transition-transform duration-300 ${
                      dragging
                        ? "rotate-0 scale-105"
                        : "rotate-[-5deg] group-hover:rotate-0"
                    }`}
                  >
                    <div className="absolute right-0 top-0 h-5 w-5 border-b border-l border-black/10 bg-[#eeeae1]" />
                    <div className="absolute left-3 top-8 h-1 w-8 rounded-full bg-black/10" />
                    <div className="absolute left-3 top-12 h-1 w-6 rounded-full bg-black/10" />
                    <div className="absolute left-3 top-16 h-1 w-8 rounded-full bg-black/10" />
                  </div>

                  {dragging ? (
                    <>
                      <p className="text-lg font-semibold tracking-tight text-[#31513d]">
                        Drop to upload
                      </p>

                      <p className="mt-2 text-sm text-[#55705d]">
                        Release your file here
                      </p>
                    </>
                  ) : file ? (
                    <>
                      <p className="max-w-full truncate text-base font-semibold">
                        {file.name}
                      </p>
                      <p className="mt-1 text-sm text-neutral-500">
                        {formatFileSize(file.size)}
                      </p>

                      <p className="mt-2 text-xs text-black/40">
                        {(file.size / 1024).toFixed(1)} KB
                        {detectedFormat
                          ? ` · ${detectedFormat.name} detected`
                          : " · Format not supported"}
                      </p>
                    </>
                 ) : mergeMode ? (
                  <>
                    <p className="text-lg font-semibold tracking-tight">
                      Drop your PDFs here
                    </p>

                    <p className="mt-2 text-sm text-black/40">
                      or click anywhere in this area to select multiple PDFs
                    </p>
                  </>
                ) : (
                  <>
                  <p className="text-lg font-semibold tracking-tight">
                    Drop your file here
                  </p>

                  <p className="mt-2 text-sm text-black/40">
                    or click anywhere in this area to browse
                  </p>
                  </>
                )}

                <div className="absolute bottom-4 left-5 right-5 flex items-center justify-between text-[10px] uppercase tracking-[0.14em] text-black/30">
                  <span>
                    {mergeMode
                      ? "PDF · PDF · PDF"
                      : rotateMode
                        ? "PDF"
                        : "PDF · DOCX · XLSX · PPTX"}
                  </span>

                  <span>Max size varies</span>
                </div>
              </div>
              
              {rotateMode && file && (
                <div className="mt-4 rounded-[18px] border border-black/10 bg-[#f7f5ef] p-4">
                  <div className="mb-3">
                    <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-black/35">
                      Rotate PDF
                    </p>
                    <p className="mt-1 truncate text-sm text-black/45">
                      {file.name}
                    </p>
                  </div>

                  <div className="grid grid-cols-3 gap-2">
                    {[90, 180, 270].map((degrees) => (
                      <button
                        key={degrees}
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          setRotation(degrees);
                        }}
                        className={
                          rotation === degrees
                            ? "rounded-xl bg-[#dce9df] px-4 py-3 text-sm font-medium text-[#31513d]"
                            : "rounded-xl border border-black/10 bg-white px-4 py-3 text-sm font-medium text-black/55 transition hover:border-black/20 hover:text-black"
                        }
                      >
                        {degrees}°
                      </button>
                    ))}
                  </div>

                  <button
                    type="button"
                    disabled={rotating}
                    onClick={(event) => {
                      event.stopPropagation();
                      rotatePDF();
                    }}
                    className="mt-4 flex w-full items-center justify-center gap-2 rounded-xl bg-[#171717] px-6 py-3 text-sm font-medium text-white transition hover:-translate-y-0.5 hover:bg-black disabled:cursor-not-allowed disabled:bg-black/20"
                  >
                    {rotating ? (
                      <>
                        <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                        Rotating...
                      </>
                    ) : (
                      "Rotate PDF →"
                    )}
                  </button>

                  {conversionError && (
                    <div
                      role="alert"
                      className="mt-3 rounded-xl border border-[#dfcfc4] bg-[#fff8f3] px-4 py-3 text-sm text-[#795e50]"
                    >
                      {conversionError}
                    </div>
                  )}

                  {downloadUrl && (
                    <div className="mt-3 rounded-xl border border-[#b9cbbd] bg-[#edf5ef] p-4">
                      <p className="text-sm font-medium text-[#31513d]">
                        Your rotated PDF is ready.
                      </p>

                      <a
                        href={downloadUrl}
                        download="rotated.pdf"
                        className="mt-3 inline-flex rounded-xl bg-[#31513d] px-4 py-2 text-xs font-medium text-white transition hover:bg-[#263f30]"
                      >
                        Download rotated PDF →
                      </a>
                    </div>
                  )}
                </div>
              )}

              {mergeMode && mergeFiles.length > 0 && (
                <div className="mt-4 rounded-[18px] border border-black/10 bg-[#f7f5ef] p-4">
                  <div className="mb-3">
                    <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-black/35">
                      PDFs to merge
                    </p>
                    <p className="mt-1 text-sm text-black/45">
                      {mergeFiles.length} PDF{mergeFiles.length === 1 ? "" : "s"} selected
                    </p>
                  </div>

                  <div className="space-y-2">
                    {mergeFiles.map((mergeFile, index) => (
                      <div
                        key={`${mergeFile.name}-${index}`}
                        className="flex items-center justify-between rounded-xl border border-black/10 bg-white px-4 py-3"
                      >
                        <span className="truncate text-sm font-medium">
                          {index + 1}. {mergeFile.name}
                        </span>

                        <span className="ml-3 shrink-0 text-xs text-black/35">
                          {formatFileSize(mergeFile.size)}
                        </span>
                      </div>
                    ))}
                  </div>

                  <button
                    type="button"
                    disabled={mergeFiles.length < 2 || merging}
                    onClick={mergePDFs}
                    className="mt-4 flex w-full items-center justify-center gap-2 rounded-xl bg-[#171717] px-6 py-3 text-sm font-medium text-white transition hover:-translate-y-0.5 hover:bg-black disabled:cursor-not-allowed disabled:bg-black/20"
                  >
                    {merging ? (
                      <>
                        <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                        Merging...
                      </>
                    ) : (
                      "Merge PDFs →"
                    )}
                  </button>

                  {mergeMode && downloadUrl && (
                    <div className="mt-6 rounded-2xl border border-black/10 bg-black/[0.02] px-5 py-4">
                      <p className="text-sm font-medium text-black">
                        Your merged PDF is ready.
                      </p>

                      <a
                        href={downloadUrl}
                        download="merged.pdf"
                        className="mt-3 inline-flex rounded-xl bg-black px-4 py-2 text-sm font-medium text-white transition hover:-translate-y-0.5 hover:bg-black/80"
                      >
                        Download merged PDF →
                      </a>
                    </div>
                  )}
                </div>
              )}

              {file && !rotateMode && (
                <>
                  <div className="mt-4 rounded-[18px] border border-black/10 bg-[#f7f5ef] p-4">
                    <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
                      <div className="flex-1">
                        <label className="mb-2 block text-[10px] font-semibold uppercase tracking-[0.15em] text-black/35">
                          Convert to
                        </label>

                        {loadingCapabilities ? (
                          <div className="rounded-xl border border-black/10 bg-white px-4 py-3 text-sm text-black/35">
                            Loading formats...
                          </div>
                        ) : capabilitiesError ? (
                          <div
                            role="alert"
                            className="rounded-xl border border-[#dfcfc4] bg-[#fff8f3] px-4 py-3 text-sm text-[#795e50]"
                          >
                            {capabilitiesError}
                          </div>
                        ) : outputFormats.length > 0 ? (
                          <select
                            value={outputFormat}
                            onChange={(event) =>
                              setOutputFormat(event.target.value)
                            }
                            className="w-full appearance-none rounded-xl border border-black/10 bg-white px-4 py-3 text-sm font-medium outline-none transition focus:border-black/30"
                          >
                            <option value="" disabled>
                              {detectedFormat
                                ? `Choose format from ${detectedFormat.extension
                                  .slice(1)
                                  .toUpperCase()}`
                                : "Select output format"}
                            </option>

                            {outputFormats.map((format) => (
                              <option
                                key={format.id}
                                value={format.id}
                              >
                                {format.name} ·{" "}
                                {format.extension.slice(1).toUpperCase()}
                              </option>
                            ))}
                          </select>
                        ) : (
                          <div className="rounded-xl border border-[#dfcfc4] bg-[#fff8f3] px-4 py-3 text-sm text-[#795e50]">
                            {isSupportedInput
                              ? "No conversion available for this file."
                              : "This file format isn't supported yet."}
                          </div>
                        )}
                      </div>

                      <button
                        type="button"
                        disabled={!outputFormat || converting}
                        onClick={convertFile}
                        className="flex min-w-[132px] items-center justify-center gap-2 rounded-xl bg-[#171717] px-6 py-3 text-sm font-medium text-white transition hover:-translate-y-0.5 hover:bg-black disabled:cursor-not-allowed disabled:bg-black/20"
                      >
                        {converting ? (
                          <>
                          <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                          Converting...
                          </>
                          ) : (
                          "Convert →"
                        )}
                      </button>
                    </div>
                  </div>

                  {conversionError && (
                    <div
                      role="alert"
                      className="mt-3 rounded-xl border border-[#dfcfc4] bg-[#fff8f3] px-4 py-3 text-sm text-[#795e50]"
                    >
                      {conversionError}
                    </div>
                  )}

                  {downloadUrl && (
                    <div className="mt-3 rounded-xl border border-[#b9cbbd] bg-[#edf5ef] p-4">
                      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                      <div>
                        <p className="text-sm font-medium text-[#31513d]">
                          Conversion complete
                        </p>

                        <p className="mt-0.5 text-xs text-[#55705d]">
                          Your converted file is ready.
                        </p>
                      </div>

                      <p className="mt-2 text-sm text-black/50">
                        {outputFormat.toUpperCase()} · Ready to download
                      </p>

                      <div className="flex shrink-0 gap-2">
                        <button
                          type="button"
                          onClick={resetWorkspace}
                          className="rounded-lg border border-[#31513d]/20 bg-white px-4 py-2 text-xs font-medium text-[#31513d] transition hover:bg-[#f7faf7]"
                        >
                          Convert another file
                        </button>

                      <a
                        href={downloadUrl}
                        download
                        className="rounded-lg bg-[#31513d] px-4 py-2 text-xs font-medium text-white transition hover:bg-[#263f30]"
                      >
                        Download
                      </a>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

            <div className="absolute -right-5 -top-7 hidden w-48 rotate-3 rounded-2xl border border-black/10 bg-[#e0eadf] p-4 shadow-[0_15px_30px_rgba(40,50,40,0.08)] sm:block">
              <div className="mb-3 text-[10px] font-semibold uppercase tracking-[0.16em] text-[#55705d]">
                Supported today
              </div>

              <div className="flex flex-wrap gap-1.5">
                {[
                  "PDF",
                  "DOCX",
                  "XLSX",
                  "PPTX",
                  "PNG",
                  "JPG",
                ].map((format) => (
                  <span
                    key={format}
                    className="rounded-md bg-white/70 px-2 py-1 text-[10px] font-medium text-[#405247]"
                  >
                    {format}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </section>

        <div className="border-t border-black/10 py-5">
          <div className="flex flex-col justify-between gap-2 text-[11px] text-black/35 sm:flex-row">
            <span>
              PaperShift — Made for Files, Not Funnels.
            </span>

            <span>
              Private by Design · Self-Hosted · Open Source
            </span>
          </div>
        </div>
      </div>
    </main>
  );
}