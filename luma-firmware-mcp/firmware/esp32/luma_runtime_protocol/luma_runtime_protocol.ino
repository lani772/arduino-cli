/*
 * LUMA runtime protocol v1 reference firmware for ESP32.
 *
 * Dependency: ArduinoJson 6.x or 7.x.
 * This sketch is read-only: it does not configure or toggle lamp GPIOs.
 * A challenge echo provides freshness, not cryptographic device authentication.
 */
#include <Arduino.h>
#include <ArduinoJson.h>
#include <esp_system.h>

#ifndef LUMA_FIRMWARE_VERSION
#define LUMA_FIRMWARE_VERSION "0.1.0"
#endif
#ifndef LUMA_PROJECT_ID
#define LUMA_PROJECT_ID "luma-esp32-reference"
#endif

static constexpr size_t kMaxRecordBytes = 4096;
static constexpr size_t kMaxChallengeBytes = 256;
static char inputLine[kMaxRecordBytes + 1];
static size_t inputLength = 0;
static bool droppingOversizedLine = false;

static String deviceId() {
  const uint64_t mac = ESP.getEfuseMac();
  char value[17];
  snprintf(value, sizeof(value), "%012llX",
           static_cast<unsigned long long>(mac & 0xFFFFFFFFFFFFULL));
  return String("esp32-") + value;
}

static void emitRecord(const char *eventName, const char *challenge) {
  DynamicJsonDocument doc(512);
  doc["protocol"] = "luma.runtime";
  doc["protocol_version"] = 1;
  doc["event"] = eventName;
  if (challenge != nullptr) {
    doc["challenge"] = challenge;
  }
  doc["firmware_version"] = LUMA_FIRMWARE_VERSION;
  doc["project_id"] = LUMA_PROJECT_ID;
  doc["device_id"] = deviceId();
  doc["uptime_ms"] = static_cast<uint32_t>(millis());

  serializeJson(doc, Serial);
  Serial.write('\n');
}

static bool validChallenge(const char *challenge) {
  if (challenge == nullptr) return false;
  const size_t length = strnlen(challenge, kMaxChallengeBytes + 1);
  if (length == 0 || length > kMaxChallengeBytes) return false;
  return challenge[0] != ' ' && challenge[length - 1] != ' ' &&
         challenge[0] != '\t' && challenge[length - 1] != '\t' &&
         challenge[0] != '\r' && challenge[length - 1] != '\r' &&
         challenge[0] != '\n' && challenge[length - 1] != '\n';
}

static void handleLine(const char *line, size_t length) {
  if (length == 0 || length > kMaxRecordBytes) return;

  DynamicJsonDocument request(512);
  const DeserializationError error = deserializeJson(request, line, length);
  if (error || !request.is<JsonObject>()) return;

  JsonObject object = request.as<JsonObject>();
  // Protocol v1 requests must contain exactly these four fields.
  if (object.size() != 4 || !object["protocol"].is<const char *>() ||
      !object["protocol_version"].is<int>() ||
      !object["command"].is<const char *>() ||
      !object["challenge"].is<const char *>()) return;

  const char *protocol = object["protocol"];
  const char *command = object["command"];
  const char *challenge = object["challenge"];
  if (strcmp(protocol, "luma.runtime") != 0 ||
      object["protocol_version"].as<int>() != 1 ||
      strcmp(command, "identify") != 0 || !validChallenge(challenge)) return;

  emitRecord("identity", challenge);
}

void setup() {
  Serial.begin(115200);
  // Allow the USB serial monitor/verifier to connect without resetting outputs.
  delay(50);
  emitRecord("ready", nullptr);
}

void loop() {
  while (Serial.available() > 0) {
    const int incoming = Serial.read();
    if (incoming < 0) return;
    const char c = static_cast<char>(incoming);

    if (c == '\n') {
      if (!droppingOversizedLine) {
        if (inputLength > 0 && inputLine[inputLength - 1] == '\r') {
          --inputLength;
        }
        inputLine[inputLength] = '\0';
        handleLine(inputLine, inputLength);
      }
      inputLength = 0;
      droppingOversizedLine = false;
      continue;
    }

    if (droppingOversizedLine) continue;
    if (inputLength >= kMaxRecordBytes) {
      // Drop the entire oversized record through its newline; never truncate it.
      droppingOversizedLine = true;
      inputLength = 0;
      continue;
    }
    inputLine[inputLength++] = c;
  }
}
