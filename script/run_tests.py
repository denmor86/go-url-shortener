
import subprocess
import os
import argparse
import tempfile
import sys

# Имя исполняемого файла тестов, должно быть добавлено в PATH
APP_NAME = "c:\develop\go\shortenertestbeta-windows-amd64.exe"

# Таймаут ожидания теста, по умолчанию 10 сек.
TIMEOUT = 10

# Путь к исходникам приложения по-умолчанию
DEFAULT_SRC_PATH = "C:\develop\go\go-url-shortener"

# Путь к исполняемому файлу приложению по-умолчанию
DEFAULT_BIN_PATH = f"{DEFAULT_SRC_PATH}\cmd\shortener\shortener.exe"

# Используемый порт по-умолчанию
DEFAULT_PORT = 8080

# Файл кэша по-умолчанию
DEFAULT_CACHE_FILE_PATH = f"{tempfile.gettempdir()}\cache.json"


def check_test_app_exist(test_bin_path):
    if not os.path.isfile(test_bin_path):
        raise FileNotFoundError(
            f'Исполняемый файл тестов {APP_NAME} не найден')


def main():
    parser = argparse.ArgumentParser(description='Запуск тестов инкрементов')
    parser.add_argument('--test_bin_path', type=str, default=APP_NAME,
                        help='Путь к приложению для тестирования')
    parser.add_argument('--increments', type=int, default=9,
                        help='Количество проверяемых инкрементов')
    parser.add_argument('--bin_path', type=str,
                        default=DEFAULT_BIN_PATH, help='Путь к приложению shortener')
    parser.add_argument('--src_path', type=str,
                        default=DEFAULT_SRC_PATH, help='Путь к исходникам приложения')
    parser.add_argument('--port', type=int, default=DEFAULT_PORT, help='Порт')
    parser.add_argument('--file_storage_path', type=str,
                        default=DEFAULT_CACHE_FILE_PATH, help='Путь к файлу кэша URLs')
    args = parser.parse_args()

    errors = False
    try:
        test_bin_path = args.test_bin_path
        check_test_app_exist(test_bin_path)

        for i in range(args.increments):
            test_verbose = "--test.v"
            current_test = f"--test.run=^TestIteration{i+1}$"
            bin_path = f"-binary-path={args.bin_path}"
            src_path = f"-source-path={args.src_path}"
            port = f"-server-port={args.port}"
            file_storage_path = f"-file-storage-path={args.file_storage_path}"
            try:
                process = subprocess.Popen([test_bin_path, test_verbose, current_test, bin_path, src_path, port, file_storage_path],
                                           stdout=sys.stdout,
                                           stderr=sys.stderr,
                                           text=True)
                print(
                    f"Тест запущен со следующими параметрами: {process.args}")
                return_code = process.wait(TIMEOUT)
                print(f"Тест завершен с кодом: {return_code}")
                if errors == False and return_code != 0:
                    errors = True
            except subprocess.TimeoutExpired:
                print(f"Ошибка. Тест не завершился за {TIMEOUT} сек.")
                errors = True
                process.kill()
            except Exception as e:
                print(f"Неизвестная ошибка: {str(e)}.")
                errors = True
                break
    except FileNotFoundError as e:
        print(f"Ошибка: {str(e)}.")
        return 1
    except Exception as e:
        print(f"Неожиданная ошибка: {str(e)}.")
        return 2

    if errors:
        print("\n Присутствуют непройденные тесты")
    else:
        print("\nВсе тесты завершены успешно")

    return 0


if __name__ == "__main__":
    exit(main())
