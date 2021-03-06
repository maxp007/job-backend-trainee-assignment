# �������� ������� �� ������� �������-��������� � ���� "������"

## ����������� ��� ������ � �������� �������������.

**��������:**

� ����� �������� ���� ����� ��������� �������������. ������ �� ��� ��� ��� ����� ����� ����������������� � �������� ������������. �� ������������� �������� ������� ������� �������������� ������ � �������� ������������ � ��������� ������. 

**������:**

���������� ����������� ����������� ��� ������ � �������� ������������� (���������� �������, �������� �������, ������� ������� �� ������������ � ������������, � ����� ����� ��������� ������� ������������). ������ ������ ������������� HTTP API � ���������/�������� �������/������ � ������� JSON. 

**�������� �������������:**

����� ������� ��������� ���������� ������ ������������ � ����������.
1. ������ �������� � ������� ������� ��������� (��� ����� visa/mastercard) ��������� ���������� ����� �� ��� ����. ������ �������� ����� �������� ��� ������ �� ������ ������������. 
2. ������������ ����� ������ � ��� �����-�� ������. ��� ����� � ��� ���� ����������� ������ ���������� ��������, ������� ����� ����������� ������ ��������� ������ � ����� ��������� ����������� �����. 
3. � ��������� ������� ����������� ���� ������������� ����������� ����������� ������ ����-����� ������ ����� ���������. �� ������ ������� ������������� ����� ����������� � �������� �� � ����������� ������ �������. 

**���������� � ����:**

1. ���� ����������: Go/PHP. �� ������ ������������� ������� �� ����� �����, �� ����������� ��� ��� �������� ������ ���.
2. ���������� � ���������� ����� ������������ �����
3. ����������� ����: MySQL ��� PostgreSQL
4. ���� ��� ������ ���� ������� �� Github � Readme ������ � ����������� �� ������� � ��������� ��������/������� (����� ������ ������� � Readme ������, ����� ����� Postman, ����� � Readme curl ������� �����������, �� ������ ����...)
5. ���� ���� �����������, ����� ���������� ����(Redis) �/��� �������(RabbitMQ, Kafka)
6. ��� ������������� �������� �� �� ��������� �������� ������� �� ���������� (� ����� ������ Readme ����� � ������� ������ ���� ������ ������ �������� � �������� �������� ���������� � ����� ������� �� �� �����)
7. ���������� ���������� � �������� �� ���������. �������������� � ��� �������������� ����������� �������� �� ���� ������� �������. ��� ������������ ����� ������������ ����� ������� ����������. ��������: � ��������� ����� curl ��� Postman.

**����� ������:**

1. ������������� docker � docker-compose ��� �������� � ������������� dev-�����.
2. ������ ��� ���������� ��������-����������� �������� ������ � ������������� ������ ���� ��� �� �������������.
3. ��� ����������� �� GO, ���-�� �� ���������� �� GO ������������. HINT: �� ������������� ��� ��� ����� ����� ������� �� Go. ��� ��������, ��� ������� :)
4. �������� unit/�������������� �����.

**�������� ������� (�������):**

����� ���������� ������� �� ������. ��������� id ������������ � ������� ������� ���������.

����� �������� ������� � �������. ��������� id ������������ � ������� ������� �������. 

����� �������� ������� �� ������������ � ������������. ��������� id ������������ � �������� ����� ������� ��������, id ������������ �������� ������ ��������� ��������, � ����� �����.

����� ��������� �������� ������� ������������. ��������� id ������������. ������ ������ � ������.

**������ �� �������:**

1. ������ ���������� � �������� ����� ���������� � ����, ���� ��� ��������� ����� �����������.
2. �� ��������� ������ �� �������� � ���� ������� ������ � �������� (������ �������� � ��). ������ � ������� ���������� ��� ������ ���������� �����. 
3. ��������� ������ � ��������� ������ ��������� �� ���������� ���������. 
4. ������ ����� � ������� �� �������������. ���������� ���� ����������� �������. � ������ ���������� ���. ������� �������� �������������� ����.
5. �������� �������� �� �����. ���������� ������������ �������� SQL ���� � ��������� ���� ����������� ������ � ��. 
6. ������ ������������ - ����� ������ ������ � ������� ����������� ������ (���������� �� �������� ��� � ��������� ��������). ���������� ������ ������� ������ � ���������� ��������� � �� ��������� �������� ����� ������ ����� ���� � �����. 
7. ������ ������� �� ��������� ������ �����.

**�������������� �������**

����� ����������� ���. �������. ��� �� �������� �������������, �� �� ���������� ���� ������������ ���� ����� ������� �����������. 

*���. ������� 1:*

����������� ��������� �������� �������� � ���� ���������� ������ � ������ � ��������� �� ����� �������. ���������� ����������� ������ ������� ������������ � �������� �� ����� ������.

������: �������� � ������ ��������� ������� ���. ��������. ������: ?currency=USD. 
���� ���� �������� ������������, �� �� ������ �������������� ������ ������������ � ����� �� ��������� ������. ������ �� �������� ����� ����� ����� ����� ������ https://exchangeratesapi.io/ ��� �� ������ ������� ��������� ���������. 

����������: ����������, ��� ������� ������ ������� �������� �� ������� � ��� ������ �����. � ������ ���� ������ ����������� ������ ���������� � ������� ������.

*���. ������� 2:*

������������ ��������, ��� �� �������� �� ��� ���� ������� (��� ���������) ��������. 

������: ���������� ������������ ����� ��������� ������ ���������� � ������������� ������ � ����� ���� ���������/������� �������� � �������. ���������� ������������� ��������� � ���������� �� ����� � ����. 




