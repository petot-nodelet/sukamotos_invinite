import { createReadStream } from "node:fs";
import { stat } from "node:fs/promises";
import { createServer } from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.dirname(fileURLToPath(import.meta.url));
const port = Number(process.env.FRONTEND_PORT || 3000);
const host = process.env.FRONTEND_HOST || "0.0.0.0";
const contentTypes = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".mjs", "text/javascript; charset=utf-8"],
  [".svg", "image/svg+xml"],
]);

createServer(async (request, response) => {
  if (!["GET", "HEAD"].includes(request.method || "")) {
    response.writeHead(405, { Allow: "GET, HEAD" }).end();
    return;
  }

  let pathname;
  try {
    pathname = decodeURIComponent(new URL(request.url || "/", "http://localhost").pathname);
  } catch {
    response.writeHead(400).end("Bad request");
    return;
  }

  const requestedPath = pathname === "/" ? "/index.html" : pathname;
  const filePath = path.resolve(root, `.${requestedPath}`);
  if (!filePath.startsWith(`${root}${path.sep}`)) {
    response.writeHead(404).end("Not found");
    return;
  }

  try {
    const fileInfo = await stat(filePath);
    if (!fileInfo.isFile()) {
      response.writeHead(404).end("Not found");
      return;
    }
    response.writeHead(200, {
      "Content-Type": contentTypes.get(path.extname(filePath)) || "application/octet-stream",
      "Content-Length": fileInfo.size,
      "X-Content-Type-Options": "nosniff",
      "Cache-Control": "no-cache",
    });
    if (request.method === "HEAD") response.end();
    else createReadStream(filePath).pipe(response);
  } catch (error) {
    if (error.code === "ENOENT" || error.code === "ENOTDIR") {
      response.writeHead(404).end("Not found");
      return;
    }
    console.error("Static frontend server error:", error);
    response.writeHead(500).end("Internal server error");
  }
}).listen(port, host, () => {
  console.log(`Frontend dev server listening on ${host}:${port}`);
});
