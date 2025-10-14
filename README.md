# CESTX - Cloud Environment for Sytem Testing and Experience (`Sistem Testi ve Deneyimi için Bulut Ortamı`)

CestX, mikroservis mimarisi kullanılarak geliştirilmiş, konteyner tabanlı bir uygulama çalıştırma ve yönetim sistemidir. Bu sistem, kullanıcıların farklı programlama dillerinde yazılmış uygulamaları kolayca çalıştırmalarını ve yönetmelerini sağlar.

## 🚀 Özellikler

- **Çoklu Dil Desteği**: Node.js, Go ve Python uygulamaları için hazır şablonlar
- **Otomatik DNS Yönetimi**: Her uygulama için otomatik alt alan adı oluşturma
- **Kaynak Yönetimi**: CPU ve bellek kullanımı izleme ve sınırlama
- **Otomatik Temizleme**: Belirli bir süre sonra kullanılmayan konteynerlerin otomatik silinmesi
- **Merkezi Loglama**: Tüm uygulamaların loglarının merkezi olarak toplanması ve analizi
- **Görev Yönetimi**: Ansible tabanlı görev yönetimi sistemi

## 🏗️ Sistem Mimarisi

Sistem aşağıdaki ana mikroservislerden oluşmaktadır:

### Ana Servisler

1. **Template Service**
   - Dockerfile ve Docker image şablonlarının yönetimi
   - MinIO, Redis ve PostgreSQL entegrasyonu
   - Şablon oluşturma ve silme işlemleri

2. **Machine Service**
   - Docker konteyner yönetimi
   - Sysbox entegrasyonu
   - Kaynak kullanımı izleme
   - Cron Job ile optimizasyon

3. **Dynoxy Service**
   - Subdomain yönetimi
   - Traefik reverse proxy entegrasyonu
   - Port yönetimi

### Yardımcı Servisler

1. **Taskmaster Service**
   - Ansible tabanlı görev yönetimi
   - Sistem kurulum ve yapılandırma görevleri

2. **Logger Service**
   - Elasticsearch ve Kibana entegrasyonu
   - Merkezi log toplama ve analizi

## 🛠️ Teknolojiler

- **Programlama Dili**: Go, Python
- **Veritabanı**: PostgreSQL
- **Cache**: Redis
- **Object Storage**: MinIO
- **Message Queue**: RabbitMQ
- **Konteynerizasyon**: Docker, Sysbox
- **Reverse Proxy**: Traefik
- **Log Yönetimi**: Elasticsearch, Kibana
- **Görev Otomasyonu**: Ansible

## 📦 Servisler Arası İletişim

- Tüm servisler RabbitMQ üzerinden mesaj kuyruğu kullanarak haberleşir
- Ortak paketler (`common`, `postgresql`, `redis`, `minio`, `rabbitmq`) tüm servislerde kullanılır
- Servis kayıt sistemi ile servislerin birbirini keşfetmesi sağlanır

## 🔒 Güvenlik

- Sysbox ile izole edilmiş konteyner ortamları
- Kaynak kullanımı sınırlamaları
- Otomatik temizleme mekanizması
- Güvenli DNS yönetimi

## 🚀 Başlangıç

Projenin kurulumu ve çalıştırılması için gerekli adımlar:

1. Gerekli servislerin kurulumu (PostgreSQL, Redis, MinIO, RabbitMQ)
2. Servislerin sırasıyla başlatılması
3. API Gateway üzerinden sisteme erişim

## 📝 Lisans

Bu proje GNU lisansı altında lisanslanmıştır. Detaylar için [LICENSE](LICENSE) dosyasına bakınız.
