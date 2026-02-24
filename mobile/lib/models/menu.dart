class MenuItem {
  final String id;
  final String name;
  final String description;
  final double price;
  final String category;

  const MenuItem({
    required this.id,
    required this.name,
    required this.description,
    required this.price,
    required this.category,
  });
}

class MenuCategory {
  final String id;
  final String name;
  final List<MenuItem> items;

  const MenuCategory({
    required this.id,
    required this.name,
    required this.items,
  });
}

class Menu {
  final String id;
  final String name;
  final String description;
  final List<MenuCategory> categories;

  const Menu({
    required this.id,
    required this.name,
    required this.description,
    required this.categories,
  });
}
